package store

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/gomodule/redigo/redis"
)

const (
	// Redis HSET which holds the issued tokens
	authCodeTokensSet = "OA2B_AC_Tokens"

	// Redis HSET which holds the issued grants until a token request is made.
	authCodeGrantSet = "OA2B_AC_Grants"

	// AuthCodeFlowID is prepended to a refresh token issued by the Authorization Code flow
	AuthCodeFlowID = "AUTHCODE"
)

// AuthCodeToken represents a token issued by the Authorization Code flow
// https://tools.ietf.org/html/rfc6749#section-4.1.3
type AuthCodeToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// authCodeTokenMeta holds the meta data of an access token
type authCodeTokenMeta struct {
	AuthGrant    string    `json:"auth_grant"`
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// NewAuthCodeToken issues new access tokens for the Authorization Code flow.
// It searches for 'code' in the Redis cache and throws errors if not found.
// If found, it checks if it has crossed is expiry limit which is 10 minutes.
// If crossed, an error is thrown.
// Else a new token is generated and returned.
// Refer RFC 6749 Section 4.1.2 (https://tools.ietf.org/html/rfc6749#section-4.1.2)
func NewAuthCodeToken(code, refreshToken, redirectURI string) (*AuthCodeToken, error) {
	// First check if such an authorization grant has been issued
	conn := cache.NewConn()
	defer cache.CloseConn(conn)

	value := code + ":" + redirectURI

	reply, err := redis.Int(conn.Do("HEXISTS", authCodeGrantSet, value))
	if err != nil {
		log.Println("NewAuthCodeToken: " + err.Error())
		return nil, err
	}

	// If 'value' is not found in the Redis cache, there are the following possibilites:
	// - A token was already issued on this authorization grant and must be revoked.
	// - It has expired and was removed by housekeep().
	// - It was never issued.
	// - the redirect URI is wrong
	if reply == 0 {
		return nil, fmt.Errorf("recycled/expired/invalid authorization grant or wrong redirect_uri")
	}

	// If found, check if it has expired since housekeeping runs only every 5 minutes
	intTime, err := redis.Int64(conn.Do("HGET", authCodeGrantSet, value))
	if err != nil {
		log.Println("NewAuthCodeToken: " + err.Error())
		return nil, err
	}

	issueTime := time.Unix(intTime, 0)
	if time.Now().Sub(issueTime) >= 10*time.Minute {
		return nil, fmt.Errorf("expired authorization grant")
	}

	// If not expired, remove it from the Redis cache since
	// we're about to issue a token for it.
	go removeAuthCodeGrant(code, redirectURI)

	var token *AuthCodeToken
	var meta *authCodeTokenMeta

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateAuthCodeToken(code)

		// Replace newly-generated refresh token with function parameter 'refreshToken'
		// if it is of length 72 since SHA-256 generates a string of length 64 and we
		// prepend it with a flow identifier of length 8. (AUTHCODE)
		if len(refreshToken) == 72 {
			token.RefreshToken = refreshToken
		}

		reply, err = redis.Int(conn.Do("HEXISTS", authCodeTokensSet, token.AccessToken))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	jsonBytes, err := json.Marshal(authCodeTokenStruct{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", authCodeTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// NewAuthCodeRefreshToken returns new token for the previously issued refresh token
// The refresh token is kept intact and can be used for future requests.
func NewAuthCodeRefreshToken(refreshToken string) (*AuthCodeToken, error) {
	code := NewAuthCodeGrant("")
	token, err := NewAuthCodeToken(code, refreshToken, "")
	if err != nil {
		return nil, err
	}

	return token, nil
}

// NewAuthCodeGrant generates a new authorization grant and adds it to a Redis cache set.
// This function takes the redirect URI as an argument, since RFC 6749 requires the same URI
// to be used in the token request as was used in the authorization grant request, if any.
// Thus, we store it along with the authorization grant in order for us to verify it against
// the one sent in the token request.
// Refer: https://tools.ietf.org/html/rfc6749#section-4.1.3
func NewAuthCodeGrant(redirectURI string) string {
	var code string
	var reply = 0
	var err error

	// In case we get a duplicate value, we iterate until we get a unique one.
	conn := cache.NewConn()
	defer cache.CloseConn(conn)
	for reply == 0 {
		code = generateNonce(20)
		value := code + ":" + redirectURI
		reply, err = redis.Int(conn.Do("HSET", authCodeGrantSet, value, time.Now().Unix()))

		if err != nil {
			log.Println(err)
		}
	}

	return code
}

// AuthCodeRefreshTokenExists checks if the refresh token exists in the Redis cache
// and returns the appropriate boolean value.
// Params:
// refreshToken: the token to look for in the cache
// invalidateIfFound: if true, the token is invalidated if found
func AuthCodeRefreshTokenExists(refreshToken string, invalidateIfFound bool) bool {
	conn := cache.NewConn()
	defer cache.CloseConn(conn)

	var token authCodeTokenStruct
	items, err := redis.ByteSlices(conn.Do("HGETALL", authCodeTokensSet))
	if err != nil {
		log.Println(err)
	}

	for i := 1; i < len(items); i += 2 {
		err := json.Unmarshal(items[i], &token)
		if err != nil {
			log.Println(err)
			break
		}

		if refreshToken == token.Token.RefreshToken {
			if invalidateIfFound {
				invalidateAuthCodeToken(token.Token.AccessToken)
			}

			return true
		}
	}

	return false
}

// VerifyAuthCodeToken checks if the token exists in the Redis cache.
// Returns true if token found, false otherwise.
func VerifyAuthCodeToken(token string) bool {
	conn := cache.NewConn()
	defer cache.CloseConn(conn)

	_, err := redis.String(conn.Do("HGET", authCodeTokensSet, token))
	return err == nil
}

func removeAuthCodeGrant(code, redirectURI string) {
	conn := cache.NewConn()
	defer cache.CloseConn(conn)
	conn.Do("HDEL", authCodeGrantSet, code+":"+redirectURI)
}

func invalidateAuthCodeToken(accessToken string) {
	conn := cache.NewConn()
	defer cache.CloseConn(conn)
	conn.Do("HDEL", authCodeTokensSet, accessToken)
}

// Generates access and refresh tokens.
// Access token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the code, time of creation and a nonce.
// Refresh token starts with the flow identifier "AUTHCODE" followed by a hex-encoded string of
// the SHA-256 hash of the concatenation of time of creation and the same nonce as above.
func generateAuthCodeToken(code string) (*AuthCodeToken, *authCodeTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := AuthCodeFlowID + hash(fmt.Sprintf("%s%s%s", code, creationTime, nonce))
	refreshToken := AuthCodeFlowID + hash(fmt.Sprintf("%s%s", creationTime, nonce))

	return &AuthCodeToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, &authCodeTokenMeta{
			AuthGrant:    code,
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}
