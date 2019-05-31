package store

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RohitAwate/OAuth2Bin/oauth2/cache"
	"github.com/gomodule/redigo/redis"
)

const ropcTokensSet = "OA2B_ROPC_Tokens"

// ROPCToken represents a token issued by the Resource Owner Password Credentials flow
// https://tools.ietf.org/html/rfc6749#section-4.3.3
type ROPCToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type ropcTokenMeta struct {
	CreationTime time.Time `json:"creation_time"`
	Nonce        string    `json:"nonce"`
}

// NewROPCToken issues new access and refresh tokens for the ROPC flow.
// It generates and stores a token and stores it along with its meta data
// in the Redis cache.
func NewROPCToken() (*ROPCToken, error) {
	conn := cache.NewConn()
	defer conn.Close()

	var token *ROPCToken
	var meta *ropcTokenMeta
	var err error
	reply := 1

	// Generates a new key if a duplicate is encountered
	for reply == 1 {
		token, meta = generateROPCToken()

		reply, err = redis.Int(conn.Do("HEXISTS", ropcTokensSet, token.AccessToken))
		if err != nil {
			log.Println(err)
			return nil, err
		}
	}

	jsonBytes, err := json.Marshal(struct {
		Token ROPCToken     `json:"token"`
		Meta  ropcTokenMeta `json:"meta"`
	}{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = conn.Do("HSET", ropcTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Generates access and refresh tokens.
// Access token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the time of creation and a nonce.
// Refresh token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the access token, the time of creation and the same nonce.
func generateROPCToken() (*ROPCToken, *ropcTokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := hash(fmt.Sprintf("%s%s", creationTime, nonce))
	refreshToken := hash(fmt.Sprintf("%s%s%s", accessToken, creationTime, nonce))

	return &ROPCToken{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		}, &ropcTokenMeta{
			CreationTime: creationTime,
			Nonce:        nonce,
		}
}
