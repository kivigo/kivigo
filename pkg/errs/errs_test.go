package errs

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBackendErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrEmptyKey", ErrEmptyKey, "key is empty"},
		{"ErrNotFound", ErrNotFound, "key not found"},
		{"ErrEmptyPrefix", ErrEmptyPrefix, "prefix is empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			require.Equal(t, tt.expected, tt.err.Error())
			require.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestKVErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrClientNotInitialized", ErrClientNotInitialized, "client is not initialized"},
		{"ErrEmptyBatch", ErrEmptyBatch, "empty batch provided"},
		{"ErrKeyNotFound", ErrKeyNotFound, "key not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			require.Equal(t, tt.expected, tt.err.Error())
			require.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestErrHealthCheckFailed(t *testing.T) {
	tests := []struct {
		name      string
		innerErr  error
		wantMsg   string
		shouldWrap bool
	}{
		{
			name:      "WithInnerError",
			innerErr:  errors.New("connection timeout"),
			wantMsg:   "health check failed: connection timeout",
			shouldWrap: true,
		},
		{
			name:      "WithComplexError",
			innerErr:  errors.New("database unreachable"),
			wantMsg:   "health check failed: database unreachable",
			shouldWrap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ErrHealthCheckFailed(tt.innerErr)
			require.Error(t, err)
			require.Equal(t, tt.wantMsg, err.Error())
			
			if tt.shouldWrap && tt.innerErr != nil {
				require.True(t, errors.Is(err, tt.innerErr))
			}
		})
	}
}

func TestErrHealthCheckFailedWithNil(t *testing.T) {
	// Test nil error case - pkg/errors.Wrap(nil, msg) returns nil
	err := ErrHealthCheckFailed(nil)
	require.Nil(t, err) // This is the actual behavior of pkg/errors.Wrap with nil
}

func TestErrorsAreDistinct(t *testing.T) {
	// Ensure all errors are distinct from each other
	allErrors := []error{
		ErrEmptyKey,
		ErrNotFound,
		ErrEmptyPrefix,
		ErrClientNotInitialized,
		ErrEmptyBatch,
		ErrKeyNotFound,
	}

	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i != j {
				require.False(t, errors.Is(err1, err2), 
					"Error %v should not be equal to %v", err1, err2)
			}
		}
	}
}

func TestErrorUnwrapping(t *testing.T) {
	innerErr := errors.New("inner error")
	wrappedErr := ErrHealthCheckFailed(innerErr)
	
	// Test that the wrapped error contains the original error
	require.True(t, errors.Is(wrappedErr, innerErr))
	
	// Note: pkg/errors.Wrap creates a different structure than standard library,
	// so we test behavior rather than exact unwrapping
	require.Contains(t, wrappedErr.Error(), innerErr.Error())
}

func TestErrorComparisons(t *testing.T) {
	// Test that backend.go and kv.go ErrNotFound/ErrKeyNotFound are different
	require.False(t, errors.Is(ErrNotFound, ErrKeyNotFound))
	require.False(t, errors.Is(ErrKeyNotFound, ErrNotFound))
	
	// Even though they have similar messages, they should be different errors
	require.NotEqual(t, ErrNotFound, ErrKeyNotFound)
}