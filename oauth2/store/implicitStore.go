package store

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"

	"github.com/gomodule/redigo/redis"
)

const implicitTokensSet = "OA2B_IG_Tokens"

// ImplicitToken represents a token issued by the Implicit Grant flow
// https://tools.ietf.org/html/rfc6749#section-4.2.2
type ImplicitToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type implicitTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// NewImplicitToken issues new access tokens for the Implicit Grant flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewImplicitToken() (*ImplicitToken, error) {
	conn := cache.NewConn()
	defer conn.Close()

	var token *ImplicitToken
	var meta *implicitTokenMeta
	var err error
	reply := 1

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateImplicitToken()

		reply, err = redis.Int(conn.Do("HEXISTS", implicitTokensSet, token.AccessToken))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	jsonBytes, err := json.Marshal(struct {
		Token ImplicitToken     `json:"token"`
		Meta  implicitTokenMeta `json:"meta"`
	}{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", implicitTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Generates an access token.
// Access token is a hex-encoded string of the SHA-256 hash of the
// concatenation of the time of creation and a nonce.
func generateImplicitToken() (*ImplicitToken, *implicitTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := hash(fmt.Sprintf("%s%s", creationTime, nonce))

	return &ImplicitToken{
			AccessToken: accessToken,
			ExpiresIn:   3600,
		}, &implicitTokenMeta{
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}
