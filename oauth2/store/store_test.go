package store

import (
	"fmt"
	"testing"
)

var strLen = 16
var cases = 1000
var generated []string

func TestRandomString(t *testing.T) {
	generated = make([]string, cases)

	for i := 0; i < cases; i++ {
		t.Run(fmt.Sprintf("#%d", i), testRandomStringFunc(strLen, i))
	}
}

func testRandomStringFunc(n, i int) func(*testing.T) {
	return func(t *testing.T) {
		newStr := randomString(n)

		for j := 0; j < i; j++ {
			if generated[j] == newStr {
				t.Errorf("Duplicate string on attempt #%d", i)
			}
		}

		generated[i] = newStr
	}
}
