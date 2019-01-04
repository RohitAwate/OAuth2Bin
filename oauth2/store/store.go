package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

var pool redis.Pool

const (
	authCodeTokensHSET = "OA2Bin_AuthCodeTokens"
)

// AuthCodeToken represents an OAuth 2.0 access token
type AuthCodeToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`

	// Used only within the store package to verify if it has expired or not
	authGrant    string
	creationTime time.Time
	randomString string
}

// NewAuthCodeToken issues new access tokens for the Authorization Code flow.
// If a recycled authorization grant is found in 'code', an HTTP 400 response is sent, and the token issued
// using that grant is revoked and an error is returned.
// Else, a new token is generated and added to the store.
// (Refer RFC 6749 Section 4.1.2 https://tools.ietf.org/html/rfc6749#section-4.1.2)
func NewAuthCodeToken(code, clientID string) (*AuthCodeToken, error) {
	creationTime := time.Now()
	exists := 1

	var token *AuthCodeToken
	var err error
	// Generates a new key if a duplicate is encountered
	for exists == 1 {
		token = generateAuthCodeToken(code, clientID, creationTime)
		token.creationTime = creationTime

		exists, err = redis.Int(pool.Get().Do("HEXISTS", authCodeTokensHSET, token.AccessToken))
		if err != nil {
			log.Println(err)
		}
	}

	jsonBytes, err := json.Marshal(token)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	_, err = pool.Get().Do("HSET", authCodeTokensHSET, token.AccessToken, string(jsonBytes))
	if err != nil {
		panic(err)
	}

	return token, nil
}

// Concatenates the code, the clientID, the string representation of the creationTime and a random string
// and generates the same code
func generateAuthCodeToken(code, clientID string, creationTime time.Time) *AuthCodeToken {
	randStr := generateRandomString(16)
	accessToken := hash(fmt.Sprintf("%s%s%s%s", code, clientID, creationTime.String(), randStr))
	refreshToken := hash(fmt.Sprintf("%s%s", creationTime, randStr))

	return &AuthCodeToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		randomString: randStr,
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
func generateRandomString(n int) string {
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

	// Initializing the connection pool with Redis
	pool = redis.Pool{
		MaxActive: 30,
		MaxIdle:   10,
		Dial: func() (redis.Conn, error) {
			var conn redis.Conn
			var err error
			if os.Getenv("REDIS_HOST") == "" && os.Getenv("REDIS_PASS") == "" && os.Getenv("REDIS_PORT") == "" {
				// Defaults to a local Redis server
				conn, err = redis.Dial("tcp", ":6379")
			} else {
				addr := fmt.Sprintf("redis://:%s@%s:%s", os.Getenv("REDIS_PASS"),
					os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
				conn, err = redis.DialURL(addr)
			}

			if err != nil {
				// Panics if connection could not be established with a Redis server
				panic(err)
			}

			return conn, nil
		},
	}
}
