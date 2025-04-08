package storage

import (
	"context"

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

func (ms *MongoStorage) Save(ctx context.Context, entityType string, id string, data interface{}) error {
	coll := ms.client.Database(ms.database).Collection(string(entityType))

	_, err := coll.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": data},
		options.UpdateOne().SetUpsert(true))

	return err
}

func (ms *MongoStorage) Load(ctx context.Context, entityType string, id string, result interface{}) error {
	coll := ms.client.Database(ms.database).Collection(string(entityType))

	return coll.FindOne(ctx, bson.M{"_id": id}).Decode(result)
}
