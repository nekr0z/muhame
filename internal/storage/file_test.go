package storage

import (
	"os"
	"path"
	"testing"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestStopAndLoad(t *testing.T) {
	t.Parallel()

	tempDir, err := os.MkdirTemp(os.TempDir(), "file_storage_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fileName := path.Join(tempDir, "test.sav")

	cfg := Config{
		Filename: fileName,
		Restore:  true,
	}

	log := zap.NewNop().Sugar()

	st := NewFileStorage(log, cfg)
	require.NoError(t, err)

	metName := "test"
	met := metrics.Counter(25)

	err = st.Update(metName, met)
	assert.NoError(t, err)

	st.Flush()

	newSt := NewFileStorage(log, cfg)
	require.NoError(t, err)

	m, err := newSt.Get(met.Type(), metName)
	assert.NoError(t, err)

	assert.Equal(t, met, m)
}
