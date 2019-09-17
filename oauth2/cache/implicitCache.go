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
	implicitTokensSet = "OA2B_IG_Tokens"

	// ImplicitFlowID is prepended to access tokens issued by the Implicit Grant flow
	ImplicitFlowID = "IMPLICIT"
)

// ImplicitToken represents a token issued by the Implicit Grant flow
// https://tools.ietf.org/html/rfc6749#section-4.2.2
type ImplicitToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// Holds the meta data of an access token
type implicitTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// Holds the token as well as its metadata.
// It is the internal representation of the token inside the Redis cache.
type internalImplicitToken struct {
	Token ImplicitToken     `json:"token"`
	Meta  implicitTokenMeta `json:"meta"`
}

// NewImplicitToken issues new access tokens for the Implicit Grant flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewImplicitToken() (*ImplicitToken, error) {
	conn := NewConn()
	defer CloseConn(conn)

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

	jsonBytes, err := json.Marshal(internalImplicitToken{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", implicitTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// VerifyImplicitToken checks if the token exists in the Redis cache.
// Returns true if token found, false otherwise.
func VerifyImplicitToken(token string) bool {
	conn := NewConn()
	defer CloseConn(conn)

	_, err := redis.String(conn.Do("HGET", implicitTokensSet, token))
	return err == nil
}

func invalidateImplicitToken(accessToken string) {
	conn := NewConn()
	defer CloseConn(conn)
	_, err := conn.Do("HDEL", implicitTokensSet, accessToken)
	if err != nil {
		log.Println(err)
	}
}

// Generates an access token.
// Access token is a hex-encoded string of the SHA-256 hash of the
// concatenation of the time of creation and a nonce.
func generateImplicitToken() (*ImplicitToken, *implicitTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := ImplicitFlowID + hash(fmt.Sprintf("%s%s", creationTime, nonce))

	return &ImplicitToken{
			AccessToken: accessToken,
			ExpiresIn:   3600,
		}, &implicitTokenMeta{
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}

// Housekeeping service for the Implicit tokens set
func implicitTokenHousekeep(conn redis.Conn) {
	var token internalImplicitToken
	var err error
	var diff time.Duration

	items, err := redis.ByteSlices(conn.Do("HGETALL", implicitTokensSet))
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
			_, err = conn.Do("HDEL", implicitTokensSet, items[i-1])
			if err != nil {
				log.Println(err)
			}
		}
	}
}
