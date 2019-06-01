package store

import "testing"

// TestImplicitFlow tests the entirety of the functions set of authCodeStore
// as they would be used by the Implicit Grant flow
func TestImplicitFlow(t *testing.T) {
	// Generating a token which would be done once the user authorizes
	// the client application
	token, err := NewImplicitToken()
	if err != nil {
		t.Fatalf("Could not generate token:\n%s\n", err)
	}

	t.Logf("Token generated: %s\n", token.AccessToken)

	// Check if token exists in the Redis cache
	res := VerifyImplicitToken(token.AccessToken)
	if !res {
		t.Fatalf("Implicit token verification failed\n")
	}

	// Remove the token from the cache
	invalidateImplicitToken(token.AccessToken)
	t.Logf("Token invalidated\n")
}
