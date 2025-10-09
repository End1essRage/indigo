package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type MongoStorage struct {
	uri string
	db  string
}

func NewMongoStorage(uri, db string) (*MongoStorage, error) {
	if err := ping(uri); err != nil {
		return nil, fmt.Errorf("ошбика подключения к mongo %w", err)
	}

	return &MongoStorage{uri: uri, db: db}, nil
}

func (fs *MongoStorage) Get(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("Ошибка подключения %w", err)
	}
	defer client.Disconnect(context.TODO())

	results := make([]Entity, 0)
	var result Entity

	col := client.Database(fs.db).Collection(collection)

	if count == 1 {
		filter := query.Bson()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := col.FindOne(ctx, filter).Decode(&result)
		if errors.Is(err, mongo.ErrNoDocuments) {
			//TODO создать тип ошибки нот фаунд
			err := fmt.Errorf("Не найдено ни одной")

			logrus.Info(err)

			return results, err
		} else if err != nil {
			err := fmt.Errorf("непридвиденная ошибка %w", err)

			logrus.Error(err)

			return results, err
		}

		results = append(results, result)
		return results, nil

	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		filter := query.Bson()

		cur, err := col.Find(ctx, filter)
		if err != nil {
			logrus.Fatal(err)
		}

		defer cur.Close(ctx)
		for cur.Next(ctx) {
			var result Entity
			if err := cur.Decode(&result); err != nil {
				logrus.Fatal(err)
			}

			results = append(results, result)
		}

		if err := cur.Err(); err != nil {
			logrus.Fatal(err)
			return results, err
		}

		return results, nil
	}
}

func (fs *MongoStorage) GetById(ctx context.Context, collection string, id string) (Entity, error) {
	return nil, nil
}

func (fs *MongoStorage) Create(ctx context.Context, collection string, entity Entity) (string, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("Ошибка подключения %w", err)
	}
	defer client.Disconnect(context.TODO())

	col := client.Database(fs.db).Collection(collection)

	result, err := col.InsertOne(ctx, entity)
	if err != nil {
		return "", err
	}

	if !result.Acknowledged {
		return "", fmt.Errorf("not acknowleged")
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (fs *MongoStorage) UpdateById(ctx context.Context, collection string, id string, entity Entity) error {
	return nil
}

func (fs *MongoStorage) Update(ctx context.Context, collection string, query QueryNode, entity Entity) (int, error) {
	return 0, nil
}

func (fs *MongoStorage) DeleteById(ctx context.Context, collection string, id string) error {
	return nil
}

func (fs *MongoStorage) Delete(ctx context.Context, collection string, query QueryNode) (int, error) {
	return 0, nil
}

func ping(uri string) error {
	client, _ := mongo.Connect(options.Client().ApplyURI(uri))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return client.Ping(ctx, readpref.Primary())
}
