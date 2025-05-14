// Package hash is used to generate hashes for the signatures.
package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

const Header = "Hashsha256" // Signature header.

// Signature generates the signature for the given message and key.
func Signature(msg []byte, key string) string {
	kb := []byte(key)
	msg = append(msg, kb...)

	sig := sha256.Sum256(msg)
	return hex.EncodeToString(sig[:])
}
