package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

const (
	defaultTableName = "kv_store"
)

type (
	Client struct {
		db        *sql.DB
		tableName string
	}

	Option struct {
		DSN       string `default:"postgres://user:password@localhost:5432/kivigo?sslmode=disable"`
		TableName string `default:"kv_store"`
	}
)

// NewOptions returns a new set of options for the PostgreSQL client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the PostgreSQL client.
// DSN: postgres://user:password@localhost:5432/kivigo?sslmode=disable
// TableName: kv_store
func DefaultOptions() Option {
	return Option{
		DSN:       "postgres://user:password@localhost:5432/kivigo?sslmode=disable",
		TableName: defaultTableName,
	}
}

// New returns a new PostgreSQL client.
func New(opt Option) (Client, error) {
	if opt.DSN == "" {
		return Client{}, fmt.Errorf("DSN is required")
	}

	if opt.TableName == "" {
		opt.TableName = defaultTableName
	}

	db, err := sql.Open("postgres", opt.DSN)
	if err != nil {
		return Client{}, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	client := Client{
		db:        db,
		tableName: opt.TableName,
	}

	// Create table if it doesn't exist
	if err := client.initTable(); err != nil {
		db.Close()
		return Client{}, fmt.Errorf("failed to initialize table: %w", err)
	}

	return client, nil
}

// initTable creates the key-value table if it doesn't exist.
func (c Client) initTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			key_name VARCHAR(255) PRIMARY KEY,
			value_data BYTEA
		);
	`, c.tableName)

	_, err := c.db.ExecContext(context.Background(), query)
	return err
}

// SetRaw stores the given raw (encoded) value under the specified key.
func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	query := fmt.Sprintf("INSERT INTO %s (key_name, value_data) VALUES ($1, $2) ON CONFLICT (key_name) DO UPDATE SET value_data = EXCLUDED.value_data", c.tableName) //nolint:gosec // table name is controlled internally
	_, err := c.db.ExecContext(ctx, query, key, value)
	return err
}

// GetRaw retrieves the raw (encoded) value stored under the specified key.
func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	// Check if key is not empty
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	query := fmt.Sprintf("SELECT value_data FROM %s WHERE key_name = $1", c.tableName) //nolint:gosec // table name is controlled internally

	var value []byte
	err := c.db.QueryRowContext(ctx, query, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return value, nil
}

// List returns a slice of all keys stored in the backend, optionally filtered by prefix.
func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	var query string
	var args []interface{}

	if prefix == "" {
		query = fmt.Sprintf("SELECT key_name FROM %s ORDER BY key_name", c.tableName)
	} else {
		query = fmt.Sprintf("SELECT key_name FROM %s WHERE key_name LIKE $1 ORDER BY key_name", c.tableName)
		args = append(args, prefix+"%")
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// Delete removes the value associated with the specified key.
func (c Client) Delete(ctx context.Context, key string) error {
	// Check if key is not empty
	if key == "" {
		return errs.ErrEmptyKey
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE key_name = $1", c.tableName) //nolint:gosec // table name is controlled internally
	_, err := c.db.ExecContext(ctx, query, key)
	return err
}

// Close closes the database connection.
func (c Client) Close() error {
	return c.db.Close()
}

// Health checks if the PostgreSQL connection is healthy.
func (c Client) Health(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// BatchGetRaw retrieves multiple keys from PostgreSQL.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	// Build placeholders for the query
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = key
	}

	query := fmt.Sprintf("SELECT key_name, value_data FROM %s WHERE key_name IN (%s)", //nolint:gosec // table name is controlled internally
		c.tableName, strings.Join(placeholders, ","))

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]byte)
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}

	// Add nil values for missing keys
	for _, key := range keys {
		if _, exists := result[key]; !exists {
			result[key] = nil
		}
	}

	return result, rows.Err()
}

// BatchSetRaw stores multiple key-value pairs in PostgreSQL.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return nil
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := fmt.Sprintf("INSERT INTO %s (key_name, value_data) VALUES ($1, $2) ON CONFLICT (key_name) DO UPDATE SET value_data = EXCLUDED.value_data", c.tableName) //nolint:gosec // table name is controlled internally
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range kv {
		if _, err := stmt.ExecContext(ctx, key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// BatchDelete removes multiple keys from PostgreSQL.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Build placeholders for the query
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = key
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE key_name IN (%s)", //nolint:gosec // table name is controlled internally
		c.tableName, strings.Join(placeholders, ","))

	_, err := c.db.ExecContext(ctx, query, args...)
	return err
}
