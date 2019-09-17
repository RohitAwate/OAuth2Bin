package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
)

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

func init() {
	// Seeding the random package
	rand.Seed(time.Now().UnixNano())
}
