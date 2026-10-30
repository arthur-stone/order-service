package repository

import (
	"context"

	"github.com/arthur-stone/order-service/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
}
