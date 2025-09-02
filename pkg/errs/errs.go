package errs

import "github.com/pkg/errors"

var (
	ErrEmptyKey              = errors.New("key is empty")
	ErrNotFound              = errors.New("key not found")
	ErrEmptyPrefix           = errors.New("prefix is empty")
	ErrOperationNotSupported = errors.New("operation not supported")
	ErrEmptyFunc             = errors.New("function is nil")
	ErrClientNotInitialized  = errors.New("client is not initialized")
	ErrEmptyBatch            = errors.New("empty batch provided")
)

var ErrHealthCheckFailed = func(err error) error {
	return errors.Wrap(err, "health check failed")
}
