//go:build unit

package cassandra

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCassandra_DefaultOptions tests the configuration defaults
func TestCassandra_DefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	require.Equal(t, []string{"127.0.0.1"}, opts.Hosts)
	require.Equal(t, "kivigo", opts.Keyspace)
	require.Equal(t, "kv", opts.Table)
}

// TestCassandra_NewOptions tests the new options constructor
func TestCassandra_NewOptions(t *testing.T) {
	opts := NewOptions()

	// NewOptions should return zero values
	require.Empty(t, opts.Hosts)
	require.Empty(t, opts.Keyspace)
	require.Empty(t, opts.Table)
}

// TestCassandra_NewWithEmptyHosts tests error handling with empty hosts
func TestCassandra_NewWithEmptyHosts(t *testing.T) {
	opts := Option{
		Hosts:    []string{},
		Keyspace: "test",
		Table:    "kv",
	}

	_, err := New(opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cassandra session")
}