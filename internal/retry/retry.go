// Package retry contains the logic for retrying functions.
package retry

import "time"

var (
	maxRetries     = 3
	backoffMul     = 2
	initialBackoff = time.Second
	maxBackoff     = 2 * time.Second
)

// OnError wraps the given function so that it is retried if the function returns a retriable error.
func OnError[T any](f func() (T, error), isRetriable func(error) bool) (T, error) {
	retries := 0
	backoff := initialBackoff

	t, err := f()
	for isRetriable(err) && retries < maxRetries {
		retries++

		backoff *= time.Duration(backoffMul)
		if backoff > maxBackoff {
			backoff = maxBackoff
		}

		time.Sleep(backoff)

		t, err = f()
	}

	return t, err
}

// Error wraps the given function so that it is retried if the function returns a retriable error.
func Error(f func() error, isRetriable func(error) bool) error {
	_, err := OnError(func() (struct{}, error) {
		return struct{}{}, f()
	}, isRetriable)
	return err
}
