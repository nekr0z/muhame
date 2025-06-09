package crypt_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nekr0z/muhame/internal/crypt"
)

var (
	update bool

	privateKey *rsa.PrivateKey

	privateKeyFile = filepath.Join("testdata", "private.key")
	publicKeyFile  = filepath.Join("testdata", "public.key")
	ciphertextFile = filepath.Join("testdata", "ciphertext")
)

func TestDecrypt(t *testing.T) {
	t.Parallel()

	message := []byte("hello world")

	t.Run("known message", func(t *testing.T) {
		t.Parallel()

		var (
			ciphertext []byte
			err        error
		)

		if update {
			ciphertext, err = crypt.Encrypt(message, &privateKey.PublicKey)
			require.NoError(t, err)

			err := os.WriteFile(ciphertextFile, ciphertext, 0644)
			require.NoError(t, err)
		} else {
			ciphertext, err = os.ReadFile(ciphertextFile)
			require.NoError(t, err)
		}

		plaintext, err := crypt.Decrypt(ciphertext, privateKey)
		assert.NoError(t, err)

		assert.Equal(t, message, plaintext)
	})

	t.Run("wrong message", func(t *testing.T) {
		t.Parallel()

		ciphertext := []byte(`-----BEGIN MESSAGE-----
v6GYThxi2yg/NgwtYfPqy/5z3oI2BhOemrysWhTMbl8=
-----END MESSAGE-----`)

		_, err := crypt.Decrypt(ciphertext, privateKey)
		assert.Error(t, err)
	})

	t.Run("not a message", func(t *testing.T) {
		t.Parallel()

		ciphertext := []byte(`hello world`)

		_, err := crypt.Decrypt(ciphertext, privateKey)
		assert.Error(t, err)
	})
}

func TestEncrypt(t *testing.T) {
	t.Parallel()

	message := []byte("gophers rule")

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ciphertext, err := crypt.Encrypt(message, &privateKey.PublicKey)
		assert.NoError(t, err)

		plaintext, err := crypt.Decrypt(ciphertext, privateKey)
		assert.NoError(t, err)

		assert.Equal(t, message, plaintext)
	})

	t.Run("nil key", func(t *testing.T) {
		t.Parallel()

		_, err := crypt.Encrypt(nil, nil)
		assert.Error(t, err)
	})

	t.Run("bad key", func(t *testing.T) {
		t.Parallel()

		_, err := crypt.Encrypt(nil, &rsa.PublicKey{
			E: 1,
		})
		assert.Error(t, err)
	})
}

func TestLoadPrivateKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "file not found",
			fileName: "notfound",
		},
		{
			name:     "not a key",
			fileName: ciphertextFile,
		},
		{
			name:     "invalid file",
			fileName: "encryption_test.go",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := crypt.LoadPrivateKey(tc.fileName)
			assert.Error(t, err)
		})
	}
}

func TestLoadPublicKey(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		pubKey, err := crypt.LoadPublicKey(publicKeyFile)
		assert.NoError(t, err)

		assert.Equal(t, privateKey.PublicKey, *pubKey)
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()

		_, err := crypt.LoadPublicKey("notfound")
		assert.Error(t, err)
	})

	t.Run("not a key", func(t *testing.T) {
		t.Parallel()

		_, err := crypt.LoadPublicKey(ciphertextFile)
		assert.Error(t, err)
	})
}

func TestMain(m *testing.M) {
	flag.BoolVar(&update, "update", false, "update keys and testdata")
	flag.Parse()

	var err error

	if update {
		privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			panic(err)
		}

		err = saveKeys(privateKey)
		if err != nil {
			panic(err)
		}
	} else {
		privateKey, err = crypt.LoadPrivateKey(privateKeyFile)
		if err != nil {
			panic(err)
		}
	}

	os.Exit(m.Run())
}

func saveKeys(key *rsa.PrivateKey) error {
	var privateKeyPEM bytes.Buffer
	err := pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err != nil {
		return err
	}

	err = os.WriteFile(privateKeyFile, privateKeyPEM.Bytes(), 0644)
	if err != nil {
		return err
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	})
	if err != nil {
		return err
	}

	return os.WriteFile(publicKeyFile, publicKeyPEM.Bytes(), 0644)
}
