// Package crypt provides cryptographic functions.
package crypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
)

// Encrypt encrypts data with public key.
func Encrypt(msg []byte, pubkey *rsa.PublicKey) ([]byte, error) {
	if pubkey == nil {
		return nil, fmt.Errorf("public key is nil")
	}

	hash := sha256.New()

	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, pubkey, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "MESSAGE",
		Bytes: ciphertext,
	}), nil
}

// Decrypt decrypts data with private key.
func Decrypt(ciphertext []byte, privkey *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.New()

	block, _ := pem.Decode(ciphertext)
	if block == nil {
		return nil, fmt.Errorf("failed to decode message")
	}

	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privkey, block.Bytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	return plaintext, nil
}
