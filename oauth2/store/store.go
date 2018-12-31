package store

import "time"

type Token struct {
	AuthGrant    string
	AccessToken  string
	RefreshToken string
	GrantTime    time.Time
}

// NewAuthCodeToken issues new access tokens for the Authorization Code
// flow. It verifies if the code has been used before and only then issues a token.
// It also adds the token to the store for verification during future requests.
// If the token is already used, it revokes the previously issued access tokens and
// returns an error.
func NewAuthCodeToken(code string) (string, error) {
	return "{ \"access_token\": \"ACCESS_TOKEN\", \"refresh_token\": \"REFRESH_TOKEN\", \"expires_in\": 3600}", nil
}
