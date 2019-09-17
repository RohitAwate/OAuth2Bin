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
	clientCredsTokensSet = "OA2B_CC_Tokens"

	// ClientCredsFlowID is prepended to access and refresh tokens issued by the Client Credentials flow
	ClientCredsFlowID = "CLICREDS"
)

// ClientCredentialsToken represents a token issued by the Resource Owner Password Credentials flow
// https://tools.ietf.org/html/rfc6749#section-4.3.3
type ClientCredentialsToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// Holds the meta data of an access token
type clientCredsTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// Holds the token as well as its metadata.
// It is the internal representation of the token inside the Redis cache.
type internalClientCredsToken struct {
	Token ClientCredentialsToken `json:"token"`
	Meta  clientCredsTokenMeta   `json:"meta"`
}

// NewClientCredsToken issues new access tokens for the Client Credentials flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewClientCredsToken() (*ClientCredentialsToken, error) {
	conn := NewConn()
	defer CloseConn(conn)

	var token *ClientCredentialsToken
	var meta *clientCredsTokenMeta
	var err error
	reply := 1

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateClientCredsToken()

		reply, err = redis.Int(conn.Do("HEXISTS", clientCredsTokensSet, token.AccessToken))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	jsonBytes, err := json.Marshal(internalClientCredsToken{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", clientCredsTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// VerifyClientCredsToken checks if the token exists in the Redis cache.
// Returns true if token found, false otherwise.
func VerifyClientCredsToken(token string) bool {
	conn := NewConn()
	defer CloseConn(conn)

	_, err := redis.String(conn.Do("HGET", clientCredsTokensSet, token))
	return err == nil
}

func invalidateClientCredsToken(accessToken string) {
	conn := NewConn()
	defer CloseConn(conn)
	_, err := conn.Do("HDEL", clientCredsTokensSet, accessToken)
	if err != nil {
		log.Println(err)
	}
}

// Generates an access token.
// Access token is a hex-encoded string of the SHA-256 hash of the
// concatenation of the time of creation and a nonce.
func generateClientCredsToken() (*ClientCredentialsToken, *clientCredsTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := ClientCredsFlowID + hash(fmt.Sprintf("%s%s", creationTime, nonce))

	return &ClientCredentialsToken{
			AccessToken: accessToken,
			ExpiresIn:   3600,
		}, &clientCredsTokenMeta{
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}

// Housekeeping service for the Client Credentials tokens set
func clientCredsTokenHousekeep(conn redis.Conn) {
	var token internalClientCredsToken
	var err error
	var diff time.Duration

	items, err := redis.ByteSlices(conn.Do("HGETALL", clientCredsTokensSet))
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
			_, err = conn.Do("HDEL", clientCredsTokensSet, items[i-1])
			if err != nil {
				log.Println(err)
			}
		}
	}
}
