package store

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	refreshToken = base64.StdEncoding.EncodeToString([]byte(code + clientID + creationTime.String()))
	accessToken = base64.StdEncoding.EncodeToString([]byte(refreshToken))
	return
}
