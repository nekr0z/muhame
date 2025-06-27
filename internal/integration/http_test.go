package integration

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nekr0z/muhame/internal/agent"
	"github.com/nekr0z/muhame/internal/server"
	"github.com/nekr0z/muhame/internal/storage"
)

func TestHTTP(t *testing.T) {
	origArgs := os.Args
	dbFile := filepath.Join("testdata", "http.sav")

	t.Cleanup(func() {
		os.Args = origArgs
	})

	os.Args = []string{
		"client",
		"-a", "localhost:23456",
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
