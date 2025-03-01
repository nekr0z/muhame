package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nekr0z/muhame/internal/metrics"
)

func TestMemStorage(t *testing.T) {
	ms := newMemStorage()
	ctx := context.Background()

	t.Run("gauge", func(t *testing.T) {
		v := metrics.Gauge(0.5)
		err := ms.Update(ctx, "test", v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("gauge_update", func(t *testing.T) {
		v := metrics.Gauge(2.4)
		err := ms.Update(ctx, "test", v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("counter", func(t *testing.T) {
		v := metrics.Counter(1)
		err := ms.Update(ctx, "test", v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, v, m)
	})

	t.Run("counter_update", func(t *testing.T) {
		v := metrics.Counter(4)
		err := ms.Update(ctx, "test", v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, m, metrics.Counter(5))
	})

	t.Run("list", func(t *testing.T) {
		ctx := context.Background()
		names, vals, err := ms.List(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, names, []string{"test", "test"})
		assert.ElementsMatch(t, vals, []metrics.Metric{metrics.Gauge(2.4), metrics.Counter(5)})
	})
}
