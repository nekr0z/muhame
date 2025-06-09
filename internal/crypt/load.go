package crypt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
)

// LoadPrivateKey reads private key.
func LoadPrivateKey(fileName string) (*rsa.PrivateKey, error) {
	block, err := readBlock(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// LoadPublicKey reads public key.
func LoadPublicKey(fileName string) (*rsa.PublicKey, error) {
	block, err := readBlock(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	key, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return key, nil
}

func readBlock(fileName string) (*pem.Block, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open key file: %w", err)
	}

	defer f.Close()

	bb, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read key: %w", err)
	}

	block, _ := pem.Decode(bb)
	if block == nil {
		return nil, fmt.Errorf("failed to decode key")
	}

	return block, nil
}
