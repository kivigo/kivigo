package mongodb

import (
	"context"
	"errors"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
		client     *mongo.Client
		database   *mongo.Database
		collection *mongo.Collection
	}

	Option struct {
		URI            string `default:"mongodb://localhost:27017"`
		DatabaseName   string `default:"kivigo"`
		CollectionName string `default:"kv"`
	}

	// Document represents the MongoDB document structure for key-value pairs
	Document struct {
		ID    string `bson:"_id"`
		Value []byte `bson:"value"`
	}
)

// NewOptions returns a new set of options for the MongoDB client.
func NewOptions() Option {
	return Option{}
}

// DefaultOptions returns the default options for the MongoDB client.
func DefaultOptions() Option {
	return Option{
		URI:            "mongodb://localhost:27017",
		DatabaseName:   "kivigo",
		CollectionName: "kv",
	}
}

// New returns a new MongoDB client.
func New(opt Option) (Client, error) {
	// Set defaults
	if opt.URI == "" {
		opt.URI = "mongodb://localhost:27017"
	}
	if opt.DatabaseName == "" {
		opt.DatabaseName = "kivigo"
	}
	if opt.CollectionName == "" {
		opt.CollectionName = "kv"
	}

	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(opt.URI))
	if err != nil {
		return Client{}, err
	}

	database := client.Database(opt.DatabaseName)
	collection := database.Collection(opt.CollectionName)

	return Client{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (c Client) Close() error {
	return c.client.Disconnect(context.Background())
}

func (c Client) SetRaw(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	doc := Document{
		ID:    key,
		Value: value,
	}

	opts := options.Replace().SetUpsert(true)
	_, err := c.collection.ReplaceOne(ctx, bson.M{"_id": key}, doc, opts)
	return err
}

func (c Client) GetRaw(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, errs.ErrEmptyKey
	}

	var doc Document
	err := c.collection.FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return doc.Value, nil
}

func (c Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return errs.ErrEmptyKey
	}

	_, err := c.collection.DeleteOne(ctx, bson.M{"_id": key})
	return err
}

func (c Client) List(ctx context.Context, prefix string) ([]string, error) {
	filter := bson.M{}
	if prefix != "" {
		filter["_id"] = bson.M{"$regex": "^" + regexp.QuoteMeta(prefix), "$options": "i"}
	}

	cursor, err := c.collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var keys []string
	for cursor.Next(ctx) {
		var doc Document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		keys = append(keys, doc.ID)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

func (c Client) Health(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

// BatchGetRaw retrieves multiple keys from MongoDB.
func (c Client) BatchGetRaw(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return nil, errors.New("empty keys slice")
	}

	filter := bson.M{"_id": bson.M{"$in": keys}}
	cursor, err := c.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string][]byte)
	for cursor.Next(ctx) {
		var doc Document
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		result[doc.ID] = doc.Value
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchSetRaw stores multiple key-value pairs in MongoDB.
func (c Client) BatchSetRaw(ctx context.Context, kv map[string][]byte) error {
	if len(kv) == 0 {
		return errors.New("empty key-value map")
	}

	// Check for empty keys
	for key := range kv {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}

	// Prepare bulk operations
	var operations []mongo.WriteModel
	for key, value := range kv {
		doc := Document{
			ID:    key,
			Value: value,
		}
		operation := mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": key}).
			SetReplacement(doc).
			SetUpsert(true)
		operations = append(operations, operation)
	}

	_, err := c.collection.BulkWrite(ctx, operations)
	return err
}

// BatchDelete removes multiple keys from MongoDB.
func (c Client) BatchDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return errors.New("empty keys slice")
	}

	// Check for empty keys
	for _, key := range keys {
		if key == "" {
			return errs.ErrEmptyKey
		}
	}

	filter := bson.M{"_id": bson.M{"$in": keys}}
	_, err := c.collection.DeleteMany(ctx, filter)
	return err
}
