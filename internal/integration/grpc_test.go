package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nekr0z/muhame/internal/agent"
	"github.com/nekr0z/muhame/internal/crypt"
	"github.com/nekr0z/muhame/internal/server"
	"github.com/nekr0z/muhame/internal/storage"
)

var (
	timeout = 8 * time.Second

	testSigningKey = "test-signing-key"
	privateKeyFile = filepath.Join("testdata", "private.key")
	publicKeyFile  = filepath.Join("testdata", "public.key")
)

func TestGRPC(t *testing.T) {
	dbFile := filepath.Join("testdata", "grpc.sav")
	origArgs := os.Args

	t.Cleanup(func() {
		os.Args = origArgs
	})

	os.Args = []string{
		"client",
		"-a", "localhost:23457",
		"-g",
		"-r", "1",
		"-p", "1",
		"-l", "1",
		"-k", testSigningKey,
		"--crypto-key", publicKeyFile,
	}
	a := agent.New()

	wg := sync.WaitGroup{}
	wg.Add(2)

	serverCtx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	os.Args = []string{
		"server",
		"-a", "localhost:23456",
		"-g", "localhost:23457",
		"-f", dbFile,
		"-i", "30",
		"-r=false",
		"-k", testSigningKey,
		"--crypto-key", privateKeyFile,
	}

	go func() {
		server.Run(serverCtx)
		t.Cleanup(func() {
			err := os.Remove(dbFile)
			require.NoError(t, err)
		})
		wg.Done()
	}()

	agentCtx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	go func() {
		a.Run(agentCtx)
		wg.Done()
	}()

	wg.Wait()

	st, err := storage.New(zaptest.NewLogger(t).Sugar(), storage.Config{
		Filename: dbFile,
		Restore:  true,
	})
	require.NoError(t, err)

	got, err := st.Get(context.Background(), "counter", "PollCount")
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestMain(m *testing.M) {
	var update bool

	flag.BoolVar(&update, "update", false, "update keys and testdata")
	flag.Parse()

	if update {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			panic(err)
		}

		err = saveKeys(privateKey)
		if err != nil {
			panic(err)
		}
	} else {
		_, err := crypt.LoadPrivateKey(privateKeyFile)
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
