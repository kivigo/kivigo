package azurecosmos

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azcosmos"

	"github.com/kivigo/kivigo/pkg/errs"
	"github.com/kivigo/kivigo/pkg/models"
)

var (
	_ models.KV           = (*Client)(nil)
	_ models.KVWithHealth = (*Client)(nil)
	_ models.KVWithBatch  = (*Client)(nil)
)

type (
	Client struct {
		client        *azcosmos.Client
		container     *azcosmos.ContainerClient
		database      string
		containerName string
	}

	Option struct {
		Endpoint      string
		Key           string
		Database      string
		Container     string
		ConnectionStr string
	}

	// CosmosItem represents a document in Cosmos DB
	CosmosItem struct {
		ID    string `json:"id"`
		PK    string `json:"pk"`    // partition key
		Value string `json:"value"` // base64 encoded value
	}
)

// NewOptions returns a new set of options for the Azure Cosmos DB client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the Azure Cosmos DB client.
func DefaultOptions() Option {
	return Option{
		Endpoint:  "https://localhost:8081",
		Key:       "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw==", // Default emulator key
		Database:  "kivigo",
		Container: "items",
	}
}

// New returns a new Azure Cosmos DB client.
func New(opt Option) (Client, error) {
	client, err := createCosmosClient(opt)
	if err != nil {
		return Client{}, fmt.Errorf("failed to create cosmos client: %w", err)
	}

	database := opt.Database
	if database == "" {
		database = "kivigo"
	}

	containerName := opt.Container
	if containerName == "" {
		containerName = "items"
	}

	c := Client{
		client:        client,
		database:      database,
		containerName: containerName,
	}

	// Initialize database and container
	if err := c.ensureDatabase(); err != nil {
		return Client{}, fmt.Errorf("failed to ensure database exists: %w", err)
	}

	if err := c.ensureContainer(); err != nil {
		return Client{}, fmt.Errorf("failed to ensure container exists: %w", err)
	}

	// Get container client
	databaseClient, err := c.client.NewDatabase(c.database)
	if err != nil {
		return Client{}, fmt.Errorf("failed to get database client: %w", err)
	}

	containerClient, err := databaseClient.NewContainer(c.containerName)
	if err != nil {
		return Client{}, fmt.Errorf("failed to get container client: %w", err)
	}

	c.container = containerClient

	return c, nil
}

// createCosmosClient creates the underlying Azure Cosmos DB client
func createCosmosClient(opt Option) (*azcosmos.Client, error) {
	endpoint := opt.Endpoint
	if endpoint == "" {
		endpoint = "https://localhost:8081"
	}

	key := opt.Key
	if key == "" {
		key = "C2y6yDjf5/R+ob0N8A7Cgv30VRDJIWEHLM+4QDU5DE2nQ9nDuVTqobD4b8mGGyPMbIZnqyMsEcaGQy67XIw/Jw=="
	}

	// Create client options
	clientOptions := &azcosmos.ClientOptions{}

	// If connecting to emulator (localhost), disable TLS verification
	if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
		clientOptions.ClientOptions = policy.ClientOptions{
			Transport: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true, //nolint:gosec // Required for emulator
					},
				},
			},
		}
	}

	if opt.ConnectionStr != "" {
		// Use connection string if provided
		return azcosmos.NewClientFromConnectionString(opt.ConnectionStr, clientOptions)
	}

	// For emulator, use the default emulator connection string format
	if strings.Contains(endpoint, "localhost") || strings.Contains(endpoint, "127.0.0.1") {
		// Build emulator connection string
		connectionStr := fmt.Sprintf("AccountEndpoint=%s;AccountKey=%s;", endpoint, key)
		return azcosmos.NewClientFromConnectionString(connectionStr, clientOptions)
	}

	// For cloud endpoints, use key-based authentication
	cred, err := azcosmos.NewKeyCredential(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create key credential: %w", err)
	}

	return azcosmos.NewClientWithKey(endpoint, cred, clientOptions)
}

func (c Client) Close() error {
	// Azure Cosmos DB client doesn't require explicit closing
	return nil
}

// ensureDatabase creates the database if it doesn't exist
func (c Client) ensureDatabase() error {
	ctx := context.Background()

	_, err := c.client.CreateDatabase(ctx, azcosmos.DatabaseProperties{
		ID: c.database,
	}, nil)
	// Ignore conflict errors (database already exists)
	if err != nil {
		var responseErr *azcore.ResponseError
		if errors.As(err, &responseErr) && responseErr.StatusCode == 409 {
			return nil // Database already exists
		}

		return err
	}

	return nil
}

// ensureContainer creates the container if it doesn't exist
func (c Client) ensureContainer() error {
	ctx := context.Background()

	databaseClient, err := c.client.NewDatabase(c.database)
	if err != nil {
		return err
	}

	containerProperties := azcosmos.ContainerProperties{
		ID: c.containerName,
		PartitionKeyDefinition: azcosmos.PartitionKeyDefinition{
			Paths: []string{"/pk"},
		},
	}

	_, err = databaseClient.CreateContainer(ctx, containerProperties, nil)
	// Ignore conflict errors (container already exists)
	if err != nil {
		var responseErr *azcore.ResponseError
		if errors.As(err, &responseErr) && responseErr.StatusCode == 409 {
			return nil // Container already exists
		}

		return err
	}

	return nil
}

// GetRaw retrieves the raw value for a given key
func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	pk := azcosmos.NewPartitionKeyString("items")
	response, err := c.container.ReadItem(ctx, pk, key, nil)
	if err != nil {
		var responseErr *azcore.ResponseError
		if errors.As(err, &responseErr) && responseErr.StatusCode == 404 {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("failed to read item: %w", err)
	}

	var item CosmosItem
	if err := json.Unmarshal(response.Value, &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return []byte(item.Value), nil
}

// SetRaw stores a raw value for a given key
func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	item := CosmosItem{
		ID:    key,
		PK:    "items", // Use a single partition key for all items to enable queries in emulator
		Value: string(value),
	}

	itemBytes, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	pk := azcosmos.NewPartitionKeyString("items")
	_, err = c.container.UpsertItem(ctx, pk, itemBytes, nil)
	if err != nil {
		return fmt.Errorf("failed to upsert item: %w", err)
	}

	return nil
}

// Delete removes a key-value pair
func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	pk := azcosmos.NewPartitionKeyString("items")
	_, err := c.container.DeleteItem(ctx, pk, key, nil)
	if err != nil {
		var responseErr *azcore.ResponseError
		if errors.As(err, &responseErr) && responseErr.StatusCode == 404 {
			return nil // Key doesn't exist, consider it successful
		}
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// List returns all keys with an optional prefix filter
func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	var query string
	opts := &azcosmos.QueryOptions{}

	if prefix == "" {
		query = "SELECT VALUE c.id FROM c"
	} else {
		query = "SELECT VALUE c.id FROM c WHERE STARTSWITH(c.id, @prefix)"
		opts.QueryParameters = []azcosmos.QueryParameter{
			{Name: "@prefix", Value: prefix},
		}
	}

	// Since all items use the same partition key "items", we can query within that partition
	pk := azcosmos.NewPartitionKeyString("items")
	queryPager := c.container.NewQueryItemsPager(query, pk, opts)

	var keys []string

	for queryPager.More() {
		response, err := queryPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to query items: %w", err)
		}

		for _, item := range response.Items {
			var id string
			if err := json.Unmarshal(item, &id); err != nil {
				continue
			}
			keys = append(keys, id)
		}
	}

	return keys, nil
}

// Health checks the health of the Azure Cosmos DB connection
func (c Client) Health(ctx context.Context) error {
	if c.client == nil || c.container == nil {
		return errs.ErrClientNotInitialized
	}

	// Try to read database info as a health check
	databaseClient, err := c.client.NewDatabase(c.database)
	if err != nil {
		return err
	}

	_, err = databaseClient.Read(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

// BatchGetRaw retrieves multiple values for given keys
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))

	// Azure Cosmos DB doesn't have a native batch get, so we'll do individual gets
	// In a real implementation, you might want to use a query for better performance
	for _, key := range keys {
		value, err := c.GetRaw(ctx, key)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				results[key] = nil // Key not found
				continue
			}
			return nil, fmt.Errorf("failed to get key %s: %w", key, err)
		}
		results[key] = value
	}

	return results, nil
}

// BatchSetRaw stores multiple key-value pairs
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	// Azure Cosmos DB doesn't have native batch operations across partition keys
	// We'll do individual sets. In production, you might want to use bulk operations
	for key, value := range kv {
		if err := c.SetRaw(ctx, key, value); err != nil {
			return fmt.Errorf("failed to set key %s: %w", key, err)
		}
	}

	return nil
}

// BatchDelete removes multiple keys
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	// Azure Cosmos DB doesn't have native batch delete across partition keys
	// We'll do individual deletes
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}
	}

	return nil
}
