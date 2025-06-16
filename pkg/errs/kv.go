package errs

import "github.com/pkg/errors"

var ErrHealthCheckFailed = func(err error) error {
	return errors.Wrap(err, "health check failed")
}

var (
	ErrClientNotInitialized = errors.New("client is not initialized")
	ErrEmptyBatch           = errors.New("empty batch provided")
	ErrKeyNotFound          = errors.New("key not found")
)
