package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStorage struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoStorage(uri, database string) (*MongoStorage, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &MongoStorage{
		client:   client,
		database: database,
	}, nil
}

func (fs *MongoStorage) Get(ctx context.Context, collection string, count int, query string) ([]Entity, error) {
	return nil, nil
}

func (fs *MongoStorage) GetById(ctx context.Context, collection string, id string) (Entity, error) {
	// Проверка контекста
	if err := ctx.Err(); err != nil {
		return Entity{}, fmt.Errorf("context error: %w", err)
	}

	errChan := make(chan error)
	resultChan := make(chan Entity)

	go func() {
		var result Entity

		if err := fs.Load(collection, id, result); err != nil {
			errChan <- fmt.Errorf("Ошбика загрзуки сущности", err)
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return Entity{}, fmt.Errorf("operation cancelled: %w", ctx.Err())
	case err := <-errChan:
		return Entity{}, err
	case data := <-resultChan:
		return data, nil
	}
}

func (fs *MongoStorage) Create(ctx context.Context, collection string, entity Entity) (string, error) {
	return "", nil
}

func (fs *MongoStorage) UpdateById(ctx context.Context, collection string, id string, entity Entity) error {
	return nil
}

func (fs *MongoStorage) Update(ctx context.Context, collection string, query string, entity Entity) (int, error) {
	return 0, nil
}

func (fs *MongoStorage) DeleteById(ctx context.Context, collection string, id string) error {
	return nil
}

func (fs *MongoStorage) Delete(ctx context.Context, collection string, query string) (int, error) {
	return 0, nil
}

func (ms *MongoStorage) Save(entityType string, data Entity) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := ms.client.Database(ms.database).Collection(string(entityType))

	_, err := coll.UpdateOne(ctx,
		bson.M{"$set": data},
		options.UpdateOne().SetUpsert(true))

	return "res.UpsertedID ", err
}

func (ms *MongoStorage) Load(entityType string, id string, result Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := ms.client.Database(ms.database).Collection(string(entityType))

	return coll.FindOne(ctx, bson.M{"_id": id}).Decode(result)
}

func (ms *MongoStorage) LoadArray(docFolder, docPath string) ([]Entity, error) {
	return nil, nil
}
