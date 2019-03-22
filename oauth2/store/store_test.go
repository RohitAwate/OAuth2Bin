package store

import (
	"fmt"
	"testing"
)

var strLen = 16
var cases = 1000
var generated []string

func TestGenerateNonce(t *testing.T) {
	generated = make([]string, cases)

	for i := 0; i < cases; i++ {
		t.Run(fmt.Sprintf("#%d", i), testGenerateNonceFunc(strLen, i))
	}
}

func testGenerateNonceFunc(n, i int) func(*testing.T) {
	return func(t *testing.T) {
		newStr := generateNonce(n)

		for j := 0; j < i; j++ {
			if generated[j] == newStr {
				t.Errorf("Duplicate string on attempt #%d", i)
			}
		}

		generated[i] = newStr
	}
}

func TestRefreshTokenExists(t *testing.T) {
	code := NewAuthCodeGrant()
	token, err := NewAuthCodeToken(code)
	if err != nil {
		t.Fatal(err)
	}

	exists := RefreshTokenExists(token.RefreshToken, true)

	if exists {
		t.Log("found refresh token")
	} else {
		removeGrant(code)
		t.Fatal("failed to find refresh token")
	}
}
