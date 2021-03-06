package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// Redis HSET which holds the issued tokens
	ropcTokensSet = "OA2B_ROPC_Tokens"

	// ROPCFlowID is prepended to access and refresh tokens issued by the ROPC flow
	ROPCFlowID = "PASSCRED"
)

// ROPCToken represents a token issued by the Resource Owner Password Credentials flow
// https://tools.ietf.org/html/rfc6749#section-4.3.3
type ROPCToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Holds the meta data of an access token
type ropcTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// Holds the token as well as its metadata.
// It is the internal representation of the token inside the Redis cache.
type internalROPCToken struct {
	Token ROPCToken     `json:"token"`
	Meta  ropcTokenMeta `json:"meta"`
}

// NewROPCToken issues new access and refresh tokens for the ROPC flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewROPCToken(refreshToken string) (*ROPCToken, error) {
	conn := NewConn()
	defer CloseConn(conn)

	var token *ROPCToken
	var meta *ropcTokenMeta
	var err error
	reply := 1

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateROPCToken()

		// Replace newly generated refresh token with function parameter 'refreshToken'
		// if it is of length 72 since SHA-256 generates a string of length 64 and we
		// prepend it with a flow identifier of length 8. (PASSCRED)
		if len(refreshToken) == 72 {
			token.RefreshToken = refreshToken
		}

		reply, err = redis.Int(conn.Do("HEXISTS", ropcTokensSet, token.AccessToken))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	jsonBytes, err := json.Marshal(internalROPCToken{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", ropcTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// NewROPCRefreshToken returns new token for the previously issued refresh token
// The refresh token is kept intact and can be used for future requests.
func NewROPCRefreshToken(refreshToken string) (*ROPCToken, error) {
	token, err := NewROPCToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// ROPCRefreshTokenExists checks if the refresh token exists in the Redis cache
// and returns the appropriate boolean value.
// Params:
// refreshToken: the token to look for in the cache
// invalidateIfFound: if true, the token is invalidated if found
func ROPCRefreshTokenExists(refreshToken string, invalidateIfFound bool) bool {
	conn := NewConn()
	defer CloseConn(conn)

	var token internalROPCToken
	items, err := redis.ByteSlices(conn.Do("HGETALL", ropcTokensSet))
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
				invalidateROPCToken(token.Token.AccessToken)
			}

			return true
		}
	}

	return false
}

// VerifyROPCToken checks if the token exists in the Redis cache.
// Returns true if token found, false otherwise.
func VerifyROPCToken(token string) bool {
	conn := NewConn()
	defer CloseConn(conn)

	_, err := redis.String(conn.Do("HGET", ropcTokensSet, token))
	return err == nil
}

func invalidateROPCToken(accessToken string) {
	conn := NewConn()
	defer CloseConn(conn)
	_, err := conn.Do("HDEL", ropcTokensSet, accessToken)
	if err != nil {
		log.Println(err)
	}
}

// Generates access and refresh tokens.
// Access token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the time of creation and a nonce.
// Refresh token starts with the flow identifier "PASSCRED" followed by the hex-encoded
// string of the SHA-256 hash of the concatenation of the access token, the time of
// creation and the same nonce.
func generateROPCToken() (*ROPCToken, *ropcTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := ROPCFlowID + hash(fmt.Sprintf("%s%s", creationTime, nonce))
	refreshToken := ROPCFlowID + hash(fmt.Sprintf("%s%s%s", accessToken, creationTime, nonce))

	return &ROPCToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, &ropcTokenMeta{
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}

// Housekeeping service for the ROPC tokens set
func ropcTokenHousekeep(conn redis.Conn) {
	var token internalROPCToken
	var err error
	var diff time.Duration

	items, err := redis.ByteSlices(conn.Do("HGETALL", ropcTokensSet))
	if err != nil {
		log.Println(err)
		return
	}

	for i := 1; i < len(items); i += 2 {
		err = json.Unmarshal(items[i], &token)
		if err != nil {
			log.Println(err)
			break
		}

		diff = time.Now().Sub(token.Meta.CreationTime)
		if diff >= time.Hour {
			_, err = conn.Do("HDEL", ropcTokensSet, items[i-1])
			if err != nil {
				log.Println(err)
			}
		}
	}
}
