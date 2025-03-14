package storage_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStopAndLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tempDir, err := os.MkdirTemp(os.TempDir(), "file_storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fileName := path.Join(tempDir, "test.sav")

	cfg := storage.Config{
		Filename: fileName,
		Restore:  true,
	}

	log := zap.NewNop().Sugar()

	st, err := storage.New(log, cfg)
	require.NoError(t, err)

	metName := "test"
	met := metrics.Counter(25)

	err = st.Update(ctx, metrics.Named{
		Name:   metName,
		Metric: met,
	})
	assert.NoError(t, err)

	st.Close()

	newSt, err := storage.New(log, cfg)
	require.NoError(t, err)

	m, err := newSt.Get(ctx, met.Type(), metName)
	assert.NoError(t, err)

	assert.Equal(t, met, m)
}
