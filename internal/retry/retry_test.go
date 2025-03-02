package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	errRetriable    = errors.New("retry me")
	errNonRetriable = errors.New("don't retry me")
)

func TestOnError_Retry(t *testing.T) {
	initialBackoff = time.Millisecond
	maxBackoff = time.Millisecond * 2

	count := 0

	c, err := OnError(func() (int, error) {
		count++
		if count < 2 {
			return -1, errRetriable
		}

		return count, nil
	}, func(err error) bool {
		return errors.Is(err, errRetriable)
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, c)
}

func TestOnError_NonRetriable(t *testing.T) {
	count := 0

	_, err := OnError(func() (int, error) {
		if count != 0 {
			t.Fatalf("should not have retried")
		}
		count++

		return count, errNonRetriable
	}, func(err error) bool {
		return errors.Is(err, errRetriable)
	})

	assert.ErrorIs(t, err, errNonRetriable)
}

func TestError_TotalTries(t *testing.T) {
	initialBackoff = time.Millisecond
	maxBackoff = time.Millisecond * 2

	count := 0

	err := Error(func() error {
		count++

		return errRetriable
	}, func(err error) bool {
		return errors.Is(err, errRetriable)
	})

	assert.ErrorIs(t, err, errRetriable)
	assert.Equal(t, 4, count) // initial + 3 additional as required
}
