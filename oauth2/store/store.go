package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

type token struct {
	AuthGrant    string
	AccessToken  string
	RefreshToken string
	CreationTime time.Time
}

// NewAuthCodeToken issues new access tokens for the Authorization Code flow.
// If a recycled authorization grant is found in 'code', an HTTP 400 response is sent, and the token issued
// using that grant is revoked and an error is returned.
// Else, a new token is generated and added to the store.
// (Refer RFC 6749 Section 4.1.2 https://tools.ietf.org/html/rfc6749#section-4.1.2)
func NewAuthCodeToken(code, clientID string) (string, error) {
	creationTime := time.Now()
	accessToken, refreshToken := generateTokens(code, clientID, creationTime)
	newToken := token{
		AuthGrant:    code,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreationTime: creationTime,
	}

	jsonStr, err := json.Marshal(struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expiry       int    `json:"expires_in"`
	}{AccessToken: newToken.AccessToken, RefreshToken: newToken.RefreshToken, Expiry: 3600})
	if err != nil {
		return "", fmt.Errorf("Could not marshal JSON")
	}

	return string(jsonStr), nil
}

// Concatenates the code, clientID and the string representation of the creationTime.
// The resultant string is Base64 encoded to generate the refreshToken which in turn is Base64 encoded
// to generate the access token.
func generateTokens(code, clientID string, creationTime time.Time) (accessToken, refreshToken string) {
	randStr := randomString(16)
	accessToken = hash(fmt.Sprintf("%s%s%s%s", code, clientID, creationTime.String(), randStr))
	refreshToken = hash(fmt.Sprintf("%s%s", creationTime, randStr))
	return
}

// Hashes the string using SHA-256
func hash(str string) string {
	hasher := sha256.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}

const src = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seeded = false

// Generates a string of given length filled with random bytes
func randomString(n int) string {
	if n < 1 {
		return ""
	}

	if !seeded {
		rand.Seed(time.Now().UnixNano())
		seeded = true
	}

	b := make([]byte, n)
	srcLen := int64(len(src))

	for i := range b {
		b[i] = src[rand.Int63()%srcLen]
	}

	return string(b)
}
