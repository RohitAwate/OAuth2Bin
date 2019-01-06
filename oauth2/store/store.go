package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var pool redis.Pool

const (
	// Redis HSET which holds the issued tokens
	authCodeTokensSet = "OA2Bin_AuthCodeTokens"

	// Redis SET which holds the issued grants until a token request is made.
	authCodeGrantSet = "OA2Bin_AuthCodeGrant"
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
func NewAuthCodeToken(code, clientID string) (*AuthCodeToken, error) {
	// First check if such an authorization grant has been issued
	reply, _ := redis.Int(pool.Get().Do("HEXISTS", authCodeGrantSet, code))

	// If not found in the Redis cache, there are three possibilites:
	// - A token was already issued on this. The previously-issued token must be revoked.
	// - It has expired and was removed by housekeep().
	// - It was never issued.
	if reply == 0 {
		return nil, fmt.Errorf("recycled, expired or invalid authorization grant")
	}

	// If found, check if it has expired since housekeeping runs only every 5 minutes
	intTime, _ := redis.Int64(pool.Get().Do("HGET", authCodeGrantSet, code))
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
		token, meta = generateAuthCodeToken(code, clientID)
		reply, _ = redis.Int(pool.Get().Do("HEXISTS", authCodeTokensSet, token.AccessToken))
	}

	jsonBytes, err := json.Marshal(struct {
		Token AuthCodeToken `json:"token"`
		Meta  tokenMeta     `json:"meta"`
	}{Token: *token, Meta: *meta})
	if err != nil {
		panic(err)
	}

	_, err = pool.Get().Do("HSET", authCodeTokensSet, token.AccessToken, string(jsonBytes))
	if err != nil {
		panic(err)
	}

	return token, nil
}

// NewAuthCodeGrant generates a new authorization grant and adds it to a Redis cache set.
func NewAuthCodeGrant() string {
	var code string
	reply := 0

	// In case we get a duplicate value, we iterate until we get a unique one.
	for reply == 0 {
		code = generateNonce(20)
		reply, _ = redis.Int(pool.Get().Do("HSET", authCodeGrantSet, code, time.Now().Unix()))
	}

	return code
}

func removeGrant(code string) {
	pool.Get().Do("HDEL", authCodeGrantSet, code)
}

func invalidateToken(accessToken string) {
	pool.Get().Do("HDEL", authCodeTokensSet, accessToken)
}

// Generates access and refresh tokens.
// Access Token is a hex-encoded string of the SHA-256 hash of the concatenation of
// the code, client ID, time of creation and a nonce.
// Refresh Token is a hex-encoded string of the SHA-256 hash of the concatenation of
// time of creation and the same nonce as above.
func generateAuthCodeToken(code, clientID string) (*AuthCodeToken, *tokenMeta) {
	nonce := generateNonce(16)
	creationTime := time.Now()

	accessToken := hash(fmt.Sprintf("%s%s%s%s", code, clientID, creationTime, nonce))
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

func tokenHousekeep(wg *sync.WaitGroup) {
	defer wg.Done()

	type tokenStruct struct {
		Token AuthCodeToken
		Meta  tokenMeta
	}

	var token tokenStruct
	var err error
	var diff time.Duration
	items, _ := redis.ByteSlices(pool.Get().Do("HGETALL", authCodeTokensSet))

	for i := 1; i < len(items); i += 2 {
		err = json.Unmarshal(items[i], &token)
		if err != nil {
			log.Println(err)
			break
		}

		diff = time.Now().Sub(token.Meta.CreationTime)
		if diff >= time.Hour {
			pool.Get().Do("HDEL", authCodeTokensSet, items[i-1])
		}
	}
}

func grantHousekeep(wg *sync.WaitGroup) {
	defer wg.Done()

	var intTime int64
	var issueTime time.Time
	grants, _ := redis.Strings(pool.Get().Do("HGETALL", authCodeGrantSet))
	for i := 0; i < len(grants); i += 2 {
		intTime, _ = strconv.ParseInt(grants[i], 10, 64)
		issueTime = time.Unix(intTime, 0)
		if time.Now().Sub(issueTime) >= time.Hour {
			pool.Get().Do("HDEL", authCodeGrantSet, grants[i])
		}
	}
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

	// Background goroutine that fires the housekeeping function every
	// 5 minutes for cleaning up expired grants and tokens.
	go func() {
		timer := time.NewTimer(5 * time.Minute)
		wg := sync.WaitGroup{}
		for {
			wg.Add(2)
			go tokenHousekeep(&wg)
			go grantHousekeep(&wg)
			wg.Wait()
			<-timer.C
		}
	}()
}
