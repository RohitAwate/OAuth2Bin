package cache

import (
	"fmt"
	"testing"
)

var strLen = 16
var iterations = 1000
var generated []string

func TestGenerateNonce(t *testing.T) {
	generated = make([]string, iterations)

	for i := 0; i < iterations; i++ {
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
