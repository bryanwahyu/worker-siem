package id

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID returns a random 128-bit hex encoded identifier.
func NewID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}
