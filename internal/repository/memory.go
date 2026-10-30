package repository

import (
	"context"
	"sync"

	"github.com/arthur-stone/order-service/internal/domain"
)

type MemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{
		orders: make(map[string]*domain.Order),
	}
}

func (r *MemoryOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copyOrder := *order
	r.orders[order.ID] = &copyOrder
	return nil
}

func (r *MemoryOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, domain.ErrOrderNotFound
	}

	copyOrder := *order
	return &copyOrder, nil
}

func (r *MemoryOrderRepository) List(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		copyOrder := *order
		all = append(all, &copyOrder)
	}

	if offset >= len(all) {
		return []*domain.Order{}, nil
	}

	end := offset + limit
	if end > len(all) || limit <= 0 {
		end = len(all)
	}

	return all[offset:end], nil
}

func (r *MemoryOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return domain.ErrOrderNotFound
	}

	copyOrder := *order
	r.orders[order.ID] = &copyOrder
	return nil
}
