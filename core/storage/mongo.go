package storage

import (
	"context"
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

//вот тут надо не принимать контекст а создавать с таймаутом ну или оборачивать в таймаут

func (ms *MongoStorage) Save(entityType string, data interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := ms.client.Database(ms.database).Collection(string(entityType))

	_, err := coll.UpdateOne(ctx,
		bson.M{"$set": data},
		options.UpdateOne().SetUpsert(true))

	return "res.UpsertedID ", err
}

func (ms *MongoStorage) Load(entityType string, id string, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := ms.client.Database(ms.database).Collection(string(entityType))

	return coll.FindOne(ctx, bson.M{"_id": id}).Decode(result)
}

func (ms *MongoStorage) LoadArray(docFolder, docPath string) ([]interface{}, error) {
	return nil, nil
}
