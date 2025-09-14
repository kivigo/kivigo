package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

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
		client    *dynamodb.Client
		tableName string
	}

	Option struct {
		Region       string
		Endpoint     string
		TableName    string
		AccessKey    string
		SecretKey    string
		SessionToken string
	}
)

// NewOptions returns a new set of options for the DynamoDB client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the DynamoDB client.
func DefaultOptions() Option {
	return Option{
		Region:    "us-east-1",
		TableName: "kivigo",
	}
}

// New returns a new DynamoDB client.
func New(opt Option) (Client, error) {
	ctx := context.Background()

	var cfg aws.Config

	var err error

	if opt.Endpoint != "" { //nolint:nestif
		// For DynamoDB Local or custom endpoint
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: opt.Endpoint}, nil
		})

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithRegion(opt.Region),
		)
		if err != nil {
			return Client{}, fmt.Errorf("failed to load config: %w", err)
		}

		if opt.AccessKey != "" && opt.SecretKey != "" {
			cfg.Credentials = credentials.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, opt.SessionToken)
		}
	} else {
		// For AWS DynamoDB
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(opt.Region))
		if err != nil {
			return Client{}, fmt.Errorf("failed to load config: %w", err)
		}

		if opt.AccessKey != "" && opt.SecretKey != "" {
			cfg.Credentials = credentials.NewStaticCredentialsProvider(opt.AccessKey, opt.SecretKey, opt.SessionToken)
		}
	}

	client := dynamodb.NewFromConfig(cfg)

	tableName := opt.TableName
	if tableName == "" {
		tableName = "kivigo"
	}

	c := Client{
		client:    client,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := c.ensureTable(ctx); err != nil {
		return Client{}, fmt.Errorf("failed to ensure table exists: %w", err)
	}

	return c, nil
}

// ensureTable creates the table if it doesn't exist
func (c Client) ensureTable(ctx context.Context) error {
	// Check if table exists
	_, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(c.tableName),
	})

	if err == nil {
		// Table exists
		return nil
	}

	// Check if error is "table not found"
	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		return err
	}

	// Table doesn't exist, create it
	_, err = c.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(c.tableName),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	if err != nil {
		return err
	}

	// For DynamoDB Local, just wait a bit instead of using the waiter
	time.Sleep(3 * time.Second)

	return nil
}

func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	_, err := c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item: map[string]types.AttributeValue{
			"PK":   &types.AttributeValueMemberS{Value: key},
			"Data": &types.AttributeValueMemberB{Value: value},
		},
	})

	return err
}

func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	result, err := c.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, errs.ErrNotFound
	}

	data, ok := result.Item["Data"]
	if !ok {
		return nil, errs.ErrNotFound
	}

	binaryData, ok := data.(*types.AttributeValueMemberB)
	if !ok {
		return nil, fmt.Errorf("unexpected data type for key %s", key)
	}

	return binaryData.Value, nil
}

func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	if prefix == "" {
		return nil, errs.ErrEmptyPrefix
	}

	var keys []string

	input := &dynamodb.ScanInput{
		TableName:            aws.String(c.tableName),
		ProjectionExpression: aws.String("PK"),
	}

	// Use scan with filter for prefix matching
	if prefix != "" {
		input.FilterExpression = aws.String("begins_with(PK, :prefix)")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":prefix": &types.AttributeValueMemberS{Value: prefix},
		}
	}

	paginator := dynamodb.NewScanPaginator(c.client, input)
	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			if pk, ok := item["PK"]; ok {
				if pkStr, ok := pk.(*types.AttributeValueMemberS); ok {
					keys = append(keys, pkStr.Value)
				}
			}
		}
	}

	return keys, nil
}

func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	_, err := c.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: key},
		},
	})

	return err
}

func (c Client) Close() error {
	// DynamoDB client doesn't require explicit closing
	return nil
}

func (c Client) Health(ctx context.Context) error {
	if c.client == nil {
		return errs.ErrClientNotInitialized
	}

	// Check if we can describe the table
	_, err := c.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(c.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}

// BatchGetRaw retrieves multiple keys from DynamoDB.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errs.ErrEmptyBatch
	}

	results := make(map[string][]byte, len(keys))

	// DynamoDB BatchGetItem can handle up to 100 items at once
	const batchSize = 100

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]
		batchResults, err := c.batchGetChunk(ctx, batchKeys)
		if err != nil {
			return nil, err
		}

		// Merge results
		for k, v := range batchResults {
			results[k] = v
		}

		// Set nil for keys that weren't found
		for _, key := range batchKeys {
			if _, found := results[key]; !found {
				results[key] = nil
			}
		}
	}

	return results, nil
}

// batchGetChunk handles a single batch of keys (helper for BatchGetRaw)
func (c Client) batchGetChunk(ctx context.Context, keys []string) (map[string][]byte, error) {
	// Build request keys
	requestItems := make([]map[string]types.AttributeValue, len(keys))
	for j, key := range keys {
		requestItems[j] = map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: key},
		}
	}

	result, err := c.client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			c.tableName: {
				Keys: requestItems,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	results := make(map[string][]byte)

	// Process results
	if items, ok := result.Responses[c.tableName]; ok {
		for _, item := range items {
			key, data := c.extractKeyAndData(item)
			if key != "" {
				results[key] = data
			}
		}
	}

	return results, nil
}

// extractKeyAndData extracts key and data from a DynamoDB item
func (c Client) extractKeyAndData(item map[string]types.AttributeValue) (string, []byte) {
	pk, ok := item["PK"]
	if !ok {
		return "", nil
	}

	pkStr, ok := pk.(*types.AttributeValueMemberS)
	if !ok {
		return "", nil
	}

	key := pkStr.Value

	data, ok := item["Data"]
	if !ok {
		return key, nil
	}

	binaryData, ok := data.(*types.AttributeValueMemberB)
	if !ok {
		return key, nil
	}

	return key, binaryData.Value
}

// BatchSetRaw sets multiple key-value pairs in DynamoDB.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errs.ErrEmptyBatch
	}

	// DynamoDB BatchWriteItem can handle up to 25 items at once
	const batchSize = 25

	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]

		// Build write requests
		writeRequests := make([]types.WriteRequest, len(batchKeys))
		for j, key := range batchKeys {
			writeRequests[j] = types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						"PK":   &types.AttributeValueMemberS{Value: key},
						"Data": &types.AttributeValueMemberB{Value: kv[key]},
					},
				},
			}
		}

		_, err := c.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				c.tableName: writeRequests,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// BatchDelete deletes multiple keys from DynamoDB.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errs.ErrEmptyBatch
	}

	// DynamoDB BatchWriteItem can handle up to 25 items at once
	const batchSize = 25

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]

		// Build delete requests
		writeRequests := make([]types.WriteRequest, len(batchKeys))
		for j, key := range batchKeys {
			writeRequests[j] = types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: key},
					},
				},
			}
		}

		_, err := c.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				c.tableName: writeRequests,
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
