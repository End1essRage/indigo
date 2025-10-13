package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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
	//подключаемся к монге
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("Ошибка подключения %s", err.Error())
		return nil, err
	}
	// закрываем подключение в конце работы
	defer client.Disconnect(context.TODO())

	//обьект для результатов
	results := make([]Entity, 0)

	//получаем коллекцию
	col := client.Database(fs.db).Collection(collection)

	//фильтр из квери
	filter := query.Bson()
	logrus.Debug(filter)

	//получаем курсор
	cur, err := col.Find(ctx, filter)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	//закрываем в конце
	defer cur.Close(ctx)
	//проходимся по дкоументам
	for cur.Next(ctx) {
		//выходим по контексту
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		//получаем и разбираем результат
		var result Entity
		if err := cur.Decode(&result); err != nil {
			logrus.Error(err)
		}

		//складываем в результат
		results = append(results, result)
	}

	//проверяем были ли ошибки
	if err := cur.Err(); err != nil {
		logrus.Error(err)
		return results, err
	}

	return results, nil

}

func (fs *MongoStorage) GetIds(ctx context.Context, collection string, count int, query QueryNode) ([]string, error) {
	//подключаемся к монге
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("Ошибка подключения %s", err.Error())
		return nil, err
	}
	// закрываем подключение в конце работы
	defer client.Disconnect(context.TODO())

	//обьект для результатов
	results := make([]string, 0)

	//получаем коллекцию
	col := client.Database(fs.db).Collection(collection)

	//фильтр из квери
	filter := query.Bson()
	logrus.Debug(filter)

	//получаем курсор
	cur, err := col.Find(ctx, filter)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	//закрываем в конце
	defer cur.Close(ctx)
	//проходимся по дкоументам
	for cur.Next(ctx) {
		//выходим по контексту
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		//получаем и разбираем результат
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}

		if err := cur.Decode(&result); err != nil {
			logrus.Error(err)
		}

		//складываем в результат
		results = append(results, result.ID.Hex())
	}

	//проверяем были ли ошибки
	if err := cur.Err(); err != nil {
		logrus.Error(err)
		return results, err
	}

	return results, nil
}

func (fs *MongoStorage) GetOne(ctx context.Context, collection string, query QueryNode) (Entity, error) {
	//подключаемся к монге
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("Ошибка подключения %s", err.Error())
		return nil, err
	}
	// закрываем подключение в конце работы
	defer client.Disconnect(context.TODO())

	//обьект для результата
	var result Entity

	//получаем коллекцию
	col := client.Database(fs.db).Collection(collection)

	//фильтр из квери
	filter := query.Bson()
	logrus.Debug(filter)

	err = col.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err := NewNotFoundError(query.ToString())

		logrus.Info(err)

		return nil, err
	} else if err != nil {
		err := fmt.Errorf("непридвиденная ошибка %w", err)

		logrus.Error(err)

		return nil, err
	}

	return result, nil
}

func (fs *MongoStorage) GetById(ctx context.Context, collection string, id string) (Entity, error) {
	return fs.GetOne(ctx, collection, &Condition{"_id", "=", id})
}

func (fs *MongoStorage) Create(ctx context.Context, collection string, entity Entity) (string, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("ошибка подключения %s", err.Error())
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
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("ошибка подключения %s", err.Error())
	}
	defer client.Disconnect(context.TODO())

	col := client.Database(fs.db).Collection(collection)

	mId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update := bson.M{"$set": entity}

	result, err := col.UpdateOne(ctx, bson.M{"_id": mId}, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = NewNotFoundError("not found")

			logrus.Info(err)
		}

		return err
	}

	if !result.Acknowledged {
		err = fmt.Errorf("не принято базой")
		return err
	}

	if result.ModifiedCount == 0 {
		err = fmt.Errorf("ни одной не обновлено")
		return err
	}

	return nil
}

func (fs *MongoStorage) Update(ctx context.Context, collection string, query QueryNode, entity Entity) (int, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("ошибка подключения %s", err.Error())
	}
	defer client.Disconnect(context.TODO())

	col := client.Database(fs.db).Collection(collection)

	update := bson.M{"$set": entity}

	result, err := col.UpdateMany(ctx, query.Bson(), update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = NewNotFoundError("not found")

			logrus.Info(err)
		}
		return 0, err
	}

	if !result.Acknowledged {
		err = fmt.Errorf("не принято базой")
		return 0, err
	}

	if result.ModifiedCount == 0 {
		err = fmt.Errorf("ни одной не обновлено")
		return 0, err
	}

	return int(result.ModifiedCount), nil
}

func (fs *MongoStorage) DeleteById(ctx context.Context, collection string, id string) error {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("ошибка подключения %s", err.Error())
	}
	defer client.Disconnect(context.TODO())

	col := client.Database(fs.db).Collection(collection)

	mId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	result, err := col.DeleteOne(ctx, bson.M{"_id": mId})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = NewNotFoundError("not found")

			logrus.Info(err)
		}
		return err
	}

	if !result.Acknowledged {
		err = fmt.Errorf("не принято базой")
		return err
	}

	if result.DeletedCount == 0 {
		err = fmt.Errorf("ни одной не удалено")
		return err
	}

	return nil
}

func (fs *MongoStorage) Delete(ctx context.Context, collection string, query QueryNode) (int, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(fs.uri))
	if err != nil {
		logrus.Errorf("ошибка подключения %s", err.Error())
	}
	defer client.Disconnect(context.TODO())

	col := client.Database(fs.db).Collection(collection)

	result, err := col.DeleteMany(ctx, query.Bson())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = NewNotFoundError("not found")

			logrus.Info(err)
		}
		return 0, err
	}

	if !result.Acknowledged {
		err = fmt.Errorf("не принято базой")
		return 0, err
	}

	if result.DeletedCount == 0 {
		err = fmt.Errorf("ни одной не удалено")
		return 0, err
	}

	return int(result.DeletedCount), nil
}

func ping(uri string) error {
	client, _ := mongo.Connect(options.Client().ApplyURI(uri))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return client.Ping(ctx, readpref.Primary())
}
