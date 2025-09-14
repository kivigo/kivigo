package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

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
		DSN       string `default:"user:password@tcp(localhost:3306)/kivigo"`
		TableName string `default:"kv_store"`
	}
)

// NewOptions returns a new set of options for the MySQL client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the MySQL client.
// DSN: user:password@tcp(localhost:3306)/kivigo
// TableName: kv_store
func DefaultOptions() Option {
	return Option{
		DSN:       "user:password@tcp(localhost:3306)/kivigo",
		TableName: defaultTableName,
	}
}

// New returns a new MySQL client.
func New(opt Option) (Client, error) {
	if opt.DSN == "" {
		return Client{}, fmt.Errorf("DSN is required")
	}

	if opt.TableName == "" {
		opt.TableName = defaultTableName
	}

	db, err := sql.Open("mysql", opt.DSN)
	if err != nil {
		return Client{}, fmt.Errorf("failed to open MySQL connection: %w", err)
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
			value_data LONGBLOB
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
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

	query := fmt.Sprintf("REPLACE INTO %s (key_name, value_data) VALUES (?, ?)", c.tableName) //nolint:gosec // table name is controlled internally
	_, err := c.db.ExecContext(ctx, query, key, value)
	return err
}

// GetRaw retrieves the raw (encoded) value stored under the specified key.
func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	// Check if key is not empty
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	query := fmt.Sprintf("SELECT value_data FROM %s WHERE key_name = ?", c.tableName) //nolint:gosec // table name is controlled internally

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
		query = fmt.Sprintf("SELECT key_name FROM %s", c.tableName)
	} else {
		query = fmt.Sprintf("SELECT key_name FROM %s WHERE key_name LIKE ?", c.tableName)
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

	query := fmt.Sprintf("DELETE FROM %s WHERE key_name = ?", c.tableName) //nolint:gosec // table name is controlled internally
	_, err := c.db.ExecContext(ctx, query, key)
	return err
}

// Close closes the database connection.
func (c Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Health checks the health of the backend connection.
func (c Client) Health(ctx context.Context) error {
	// Check if the client is nil
	if c.db == nil {
		return errs.ErrClientNotInitialized
	}

	// Ping the MySQL server to check health
	if err := c.db.PingContext(ctx); err != nil {
		return err
	}

	return nil
}

// BatchGetRaw retrieves multiple raw values for the given keys.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	// Check if keys slice is not empty
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))

	// Build query with placeholders
	placeholders := strings.Repeat("?,", len(keys))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf("SELECT key_name, value_data FROM %s WHERE key_name IN (%s)", c.tableName, placeholders) //nolint:gosec // table name is controlled internally

	// Convert keys to interface{} slice for query args
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize all keys to nil (not found)
	for _, key := range keys {
		results[key] = nil
	}

	// Fill in found values
	for rows.Next() {
		var key string
		var value []byte
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		results[key] = value
	}

	return results, rows.Err()
}

// validateKeys checks if any key in the slice is empty.
func validateKeys(keys []string) error {
	for _, key := range keys {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}
	return nil
}

// validateKeyValuePairs checks if any key in the map is empty.
func validateKeyValuePairs(kv map[string][]byte) error {
	for key := range kv {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}
	return nil
}

// executeTransaction runs a transaction with automatic rollback handling.
func (c Client) executeTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var committed bool
	defer func() {
		if !committed {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error if needed, but don't override the main error
			}
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// BatchSetRaw stores multiple key-value pairs atomically.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	if err := validateKeyValuePairs(kv); err != nil {
		return err
	}

	return c.executeTransaction(ctx, func(tx *sql.Tx) error {
		query := fmt.Sprintf("REPLACE INTO %s (key_name, value_data) VALUES (?, ?)", c.tableName) //nolint:gosec // table name is controlled internally
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
		return nil
	})
}

// BatchDelete removes multiple keys atomically.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	if err := validateKeys(keys); err != nil {
		return err
	}

	return c.executeTransaction(ctx, func(tx *sql.Tx) error {
		query := fmt.Sprintf("DELETE FROM %s WHERE key_name = ?", c.tableName) //nolint:gosec // table name is controlled internally
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, key := range keys {
			if _, err := stmt.ExecContext(ctx, key); err != nil {
				return err
			}
		}
		return nil
	})
}
