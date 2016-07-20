package util

import (
	"math/rand"
	"time"
)

const (
	validRunes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	idxBits    = 6
	idxMask    = 1<<idxBits - 1
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SessionID Generates a random session ID. This function is taken from tgomaild2
func SessionID(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & idxMask); idx < len(validRunes) {
			b[i] = validRunes[idx]
			i++
		}
	}
	return string(b)
}
