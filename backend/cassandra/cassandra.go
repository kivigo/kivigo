package cassandra

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gocql/gocql"

	"github.com/azrod/kivigo/pkg/errs"
	"github.com/azrod/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

type (
	Client struct {
		session  *gocql.Session
		keyspace string
		table    string
	}

	Option struct {
		Hosts    []string `default:"[\"127.0.0.1\"]"`
		Keyspace string   `default:"kivigo"`
		Table    string   `default:"kv"`
	}
)

// NewOptions returns a new set of options for the Cassandra client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the Cassandra client.
func DefaultOptions() Option {
	return Option{
		Hosts:    []string{"127.0.0.1"},
		Keyspace: "kivigo",
		Table:    "kv",
	}
}

// New returns a new Cassandra client.
func New(opt Option) (Client, error) {
	cluster := gocql.NewCluster(opt.Hosts...)
	cluster.Keyspace = opt.Keyspace
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return Client{}, fmt.Errorf("could not create cassandra session: %w", err)
	}

	client := Client{
		session:  session,
		keyspace: opt.Keyspace,
		table:    opt.Table,
	}

	// Create keyspace and table if they don't exist
	if err := client.initSchema(); err != nil {
		session.Close()
		return Client{}, fmt.Errorf("could not initialize cassandra schema: %w", err)
	}

	return client, nil
}

func (c Client) Close() error {
	if c.session != nil {
		c.session.Close()
	}
	return nil
}

// initSchema creates the keyspace and table if they don't exist.
func (c Client) initSchema() error {
	// Create keyspace
	createKeyspaceQuery := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH REPLICATION = {
			'class': 'SimpleStrategy',
			'replication_factor': 1
		}
	`, c.keyspace)

	if err := c.session.Query(createKeyspaceQuery).Exec(); err != nil {
		return fmt.Errorf("could not create keyspace: %w", err)
	}

	// Create table
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
			key text PRIMARY KEY,
			value blob
		)
	`, c.keyspace, c.table)

	if err := c.session.Query(createTableQuery).Exec(); err != nil {
		return fmt.Errorf("could not create table: %w", err)
	}

	return nil
}

// SetRaw stores the given raw (encoded) value under the specified key.
func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	query := fmt.Sprintf("INSERT INTO %s.%s (key, value) VALUES (?, ?)", c.keyspace, c.table)
	return c.session.Query(query, key, value).WithContext(ctx).Exec()
}

// GetRaw retrieves the raw (encoded) value stored under the specified key.
func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	var value []byte
	query := fmt.Sprintf("SELECT value FROM %s.%s WHERE key = ?", c.keyspace, c.table)
	err := c.session.Query(query, key).WithContext(ctx).Scan(&value)
	if err != nil {
		if errors.Is(err, gocql.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return value, nil
}

// Delete removes the value associated with the specified key.
func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE key = ?", c.keyspace, c.table)
	return c.session.Query(query, key).WithContext(ctx).Exec()
}

// List returns a slice of all keys stored in the backend, optionally filtered by prefix.
func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	if prefix == "" {
		return nil, errs.ErrEmptyPrefix
	}

	var keys []string
	query := fmt.Sprintf("SELECT key FROM %s.%s", c.keyspace, c.table)

	iter := c.session.Query(query).WithContext(ctx).Iter()
	defer iter.Close()

	var key string
	for iter.Scan(&key) {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}

	return keys, nil
}

// Health checks the health of the backend connection.
func (c Client) Health(ctx context.Context) error {
	query := fmt.Sprintf("SELECT key FROM %s.%s LIMIT 1", c.keyspace, c.table)
	return c.session.Query(query).WithContext(ctx).Exec()
}

// BatchGetRaw retrieves multiple raw values for the given keys.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	// Check for empty keys
	for _, key := range keys {
		if key == "" {
			return nil, errs.ErrEmptyKey
		}
	}

	results := make(map[string][]byte, len(keys))

	// Initialize all keys to nil (not found)
	for _, key := range keys {
		results[key] = nil
	}

	// For Cassandra, we need to execute multiple queries since it doesn't support IN queries efficiently
	// for large sets of keys. We'll use concurrent queries for better performance.
	query := fmt.Sprintf("SELECT value FROM %s.%s WHERE key = ?", c.keyspace, c.table)

	for _, key := range keys {
		var value []byte
		err := c.session.Query(query, key).WithContext(ctx).Scan(&value)
		if err == nil {
			results[key] = value
		} else if !errors.Is(err, gocql.ErrNotFound) {
			return nil, err
		}
		// If err == gocql.ErrNotFound, we keep results[key] = nil
	}

	return results, nil
}

// BatchSetRaw stores multiple key-value pairs.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	// Check for empty keys
	for key := range kv {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}

	// Use Cassandra batch for atomic writes
	batch := c.session.NewBatch(gocql.LoggedBatch)
	query := fmt.Sprintf("INSERT INTO %s.%s (key, value) VALUES (?, ?)", c.keyspace, c.table)

	for key, value := range kv {
		batch.Query(query, key, value)
	}

	return c.session.ExecuteBatch(batch.WithContext(ctx))
}

// BatchDelete removes multiple keys.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	// Check for empty keys
	for _, key := range keys {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}

	// Use Cassandra batch for atomic deletes
	batch := c.session.NewBatch(gocql.LoggedBatch)
	query := fmt.Sprintf("DELETE FROM %s.%s WHERE key = ?", c.keyspace, c.table)

	for _, key := range keys {
		batch.Query(query, key)
	}

	return c.session.ExecuteBatch(batch.WithContext(ctx))
}
