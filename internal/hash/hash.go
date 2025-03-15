package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

const Header = "Hashsha256"

func Signature(msg []byte, key string) string {
	kb := []byte(key)
	msg = append(msg, kb...)

	sig := sha256.Sum256(msg)
	return hex.EncodeToString(sig[:])
}
