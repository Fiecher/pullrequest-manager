package repositories

import "context"

type Repository[E any, ID any] interface {
	Create(ctx context.Context, entity *E) error
	FindByID(ctx context.Context, id ID) (*E, error)
	FindAll(ctx context.Context) ([]*E, error)
	Update(ctx context.Context, entity *E) error
	DeleteByID(ctx context.Context, id ID) error
}
