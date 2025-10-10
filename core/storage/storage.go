package storage

import "context"

type Entity map[string]interface{}

func NewEntity() Entity {
	return make(Entity)
}

type Storage interface {
	//0 count - all
	Get(ctx context.Context, collection string, count int, query QueryNode) ([]Entity, error)
	GetIds(ctx context.Context, collection string, count int, query QueryNode) ([]string, error)
	GetOne(ctx context.Context, collection string, query QueryNode) (Entity, error)
	GetById(ctx context.Context, collection string, id string) (Entity, error)

	Create(ctx context.Context, collection string, entity Entity) (string, error)

	UpdateById(ctx context.Context, collection string, id string, entity Entity) error
	Update(ctx context.Context, collection string, query QueryNode, entity Entity) (int, error)

	DeleteById(ctx context.Context, collection string, id string) error
	Delete(ctx context.Context, collection string, query QueryNode) (int, error)
}
