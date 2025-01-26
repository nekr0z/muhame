package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
)

func TestMemStorage(t *testing.T) {
	ms := storage.NewMemStorage()

	t.Run("gauge", func(t *testing.T) {
		v := metrics.Gauge(0.5)
		err := ms.Update("test", v)
		assert.NoError(t, err)

		m, err := ms.Get("gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("gauge_update", func(t *testing.T) {
		v := metrics.Gauge(2.4)
		err := ms.Update("test", v)
		assert.NoError(t, err)

		m, err := ms.Get("gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("counter", func(t *testing.T) {
		v := metrics.Counter(1)
		err := ms.Update("test", v)
		assert.NoError(t, err)

		m, err := ms.Get("counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("counter_update", func(t *testing.T) {
		v := metrics.Counter(4)
		err := ms.Update("test", v)
		assert.NoError(t, err)

		m, err := ms.Get("counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, m, metrics.Counter(5))
	})

	t.Run("list", func(t *testing.T) {
		names, vals, err := ms.List()
		assert.NoError(t, err)
		assert.ElementsMatch(t, names, []string{"test", "test"})
		assert.ElementsMatch(t, vals, []metrics.Metric{metrics.Gauge(2.4), metrics.Counter(5)})
	})
}
