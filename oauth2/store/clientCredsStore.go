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

type clientCredsTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// NewClientCredsToken issues new access tokens for the Client Credentials flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewClientCredsToken() (*ClientCredentialsToken, error) {
	conn := cache.NewConn()
	defer conn.Close()

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

	jsonBytes, err := json.Marshal(struct {
		Token ClientCredentialsToken `json:"token"`
		Meta  clientCredsTokenMeta   `json:"meta"`
	}{Token: *token, Meta: *meta})
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
	conn := cache.NewConn()
	defer conn.Close()

	_, err := redis.String(conn.Do("HGET", clientCredsTokensSet, token))
	return err == nil
}

func invalidateClientCredsToken(accessToken string) {
	conn := cache.NewConn()
	defer conn.Close()
	conn.Do("HDEL", clientCredsTokensSet, accessToken)
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
