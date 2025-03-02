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
		v := metrics.Named{
			Name:   "test",
			Metric: metrics.Gauge(0.5),
		}
		err := ms.Update(ctx, v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v.Metric, m)
	})

	t.Run("gauge_update", func(t *testing.T) {
		v := metrics.Named{
			Name:   "test",
			Metric: metrics.Gauge(2.4),
		}
		err := ms.Update(ctx, v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "gauge", "test")
		assert.NoError(t, err)
		assert.Equal(t, v.Metric, m)
	})

	t.Run("counter", func(t *testing.T) {
		v := metrics.Named{
			Name:   "test",
			Metric: metrics.Counter(1),
		}
		err := ms.Update(ctx, v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, v.Metric, m)
	})

	t.Run("counter_update", func(t *testing.T) {
		v := metrics.Named{
			Name:   "test",
			Metric: metrics.Counter(4),
		}
		err := ms.Update(ctx, v)
		assert.NoError(t, err)

		m, err := ms.Get(ctx, "counter", "test")
		assert.NoError(t, err)
		assert.Equal(t, m, metrics.Counter(5))
	})

	t.Run("list", func(t *testing.T) {
		ctx := context.Background()
		ms, err := ms.List(ctx)
		assert.NoError(t, err)
		assert.ElementsMatch(t, ms, []metrics.Named{
			{
				Name:   "test",
				Metric: metrics.Gauge(2.4),
			},
			{
				Name:   "test",
				Metric: metrics.Counter(5),
			},
		})
	})
}
