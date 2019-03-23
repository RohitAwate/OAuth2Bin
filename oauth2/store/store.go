package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/gomodule/redigo/redis"
)

const (
	// Redis HSET which holds the issued tokens
	authCodeTokensSet = "OA2B_AC_Tokens"

	// Redis SET which holds the issued grants until a token request is made.
	authCodeGrantSet = "OA2B_AC_Grants"
)

// tokenMeta holds the meta data of an access token
type tokenMeta struct {
	AuthGrant    string    `json:"auth_grant"`
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// AuthCodeToken represents an OAuth 2.0 access token
type AuthCodeToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// NewAuthCodeToken issues new access tokens for the Authorization Code flow.
// It searches for 'code' in the Redis cache and throws errors if not found.
// If found, it checks if it has crossed is expiry limit which is 10 minutes.
// If crossed, an error is thrown.
// Else a new token is generated and returned.
// Refer RFC 6749 Section 4.1.2 (https://tools.ietf.org/html/rfc6749#section-4.1.2)
func NewAuthCodeToken(code string) (*AuthCodeToken, error) {
	// First check if such an authorization grant has been issued
	conn := cache.NewConn()
	defer conn.Close()
	reply, _ := redis.Int(conn.Do("HEXISTS", authCodeGrantSet, code))

	// If not found in the Redis cache, there are three possibilites:
	// - A token was already issued on this. The previously-issued token must be revoked.
	// - It has expired and was removed by housekeep().
	// - It was never issued.
	if reply == 0 {
		return nil, fmt.Errorf("recycled, expired or invalid authorization grant")
	}

	// If found, check if it has expired since housekeeping runs only every 5 minutes
	intTime, _ := redis.Int64(conn.Do("HGET", authCodeGrantSet, code))
	issueTime := time.Unix(intTime, 0)
	if time.Now().Sub(issueTime) >= 10*time.Minute {
		return nil, fmt.Errorf("expired authorization grant")
	}

	// If not expired, remove it from the Redis cache since
	// we're about to issue a token for it
	removeGrant(code)

	var token *AuthCodeToken
	var meta *tokenMeta
	var err error

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateAuthCodeToken(code)
		reply, _ = redis.Int(conn.Do("HEXISTS", authCodeTokensSet, token.AccessToken))
	}

	jsonBytes, err := json.Marshal(struct {
		Token AuthCodeToken `json:"token"`
		Meta  tokenMeta     `json:"meta"`
	}{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", authCodeTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// NewRefreshToken returns
func NewRefreshToken(refreshToken string) (*AuthCodeToken, error) {
	code := NewAuthCodeGrant()
	token, err := NewAuthCodeToken(code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// NewAuthCodeGrant generates a new authorization grant and adds it to a Redis cache set.
func NewAuthCodeGrant() string {
	var code string
	reply := 0

	// In case we get a duplicate value, we iterate until we get a unique one.
	conn := cache.NewConn()
	defer conn.Close()
	for reply == 0 {
		code = generateNonce(20)
		reply, _ = redis.Int(conn.Do("HSET", authCodeGrantSet, code, time.Now().Unix()))
	}

	return code
}

// RefreshTokenExists checks if the refresh token exists in the Redis cache
// and returns the appropriate boolean value.
// Params:
// refreshToken: the token to look for in the cache
// invalidateIfFound: if true, the token is invalidated if found
func RefreshTokenExists(refreshToken string, invalidateIfFound bool) bool {
	conn := cache.NewConn()
	defer conn.Close()

	var token tokenStruct
	items, _ := redis.ByteSlices(conn.Do("HGETALL", authCodeTokensSet))

	for i := 1; i < len(items); i += 2 {
		err := json.Unmarshal(items[i], &token)
		if err != nil {
			log.Println(err)
			break
		}

		if refreshToken == token.Token.RefreshToken {
			if invalidateIfFound {
				invalidateToken(token.Token.AccessToken)
			}

			return true
		}
	}

	return false
}

func removeGrant(code string) {
	conn := cache.NewConn()
	defer conn.Close()
	conn.Do("HDEL", authCodeGrantSet, code)
}

func invalidateToken(accessToken string) {
	conn := cache.NewConn()
	defer conn.Close()
	conn.Do("HDEL", authCodeTokensSet, accessToken)
}

// Generates access and refresh tokens.
// Access Token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the code, time of creation and a nonce.
// Refresh Token is a hex-encoded string of the SHA-256 hash of the concatenation of
// time of creation and the same nonce as above.
func generateAuthCodeToken(code string) (*AuthCodeToken, *tokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := hash(fmt.Sprintf("%s%s%s", code, creationTime, nonce))
	refreshToken := hash(fmt.Sprintf("%s%s", creationTime, nonce))

	return &AuthCodeToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, &tokenMeta{
			AuthGrant:    code,
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}

// Hashes the string using SHA-256
func hash(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

const src = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Generates a string of given length filled with random bytes
func generateNonce(n int) string {
	if n < 1 {
		return ""
	}

	b := make([]byte, n)
	srcLen := int64(len(src))

	for i := range b {
		b[i] = src[rand.Int63()%srcLen]
	}

	return string(b)
}

func init() {
	// Seeding the random package
	rand.Seed(time.Now().UnixNano())
}
