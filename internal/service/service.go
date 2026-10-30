package service

import (
	"context"
	"fmt"

	"github.com/arthur-stone/order-service/internal/domain"
	"github.com/arthur-stone/order-service/internal/repository"
)

type OrderService interface {
	CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error)
	PayOrder(ctx context.Context, id string) (*domain.Order, error)
	CancelOrder(ctx context.Context, id string) (*domain.Order, error)
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) CreateOrder(ctx context.Context, customerID string, items []domain.OrderItem) (*domain.Order, error) {
	order, err := domain.NewOrder(customerID, items)
	if err != nil {
		return nil, fmt.Errorf("domain.NewOrder: %w", err)
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("repo.Save: %w", err)
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repo.GetByID: %w", err)
	}
	return order, nil
}

func (s *orderService) ListOrders(ctx context.Context, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repo.List: %w", err)
	}
	return orders, nil
}

func (s *orderService) PayOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repo.GetByID: %w", err)
	}

	if err := order.Pay(); err != nil {
		return nil, fmt.Errorf("order.Pay: %w", err)
	}

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("repo.Update: %w", err)
	}

	return order, nil
}

func (s *orderService) CancelOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("repo.GetByID: %w", err)
	}

	if err := order.Cancel(); err != nil {
		return nil, fmt.Errorf("order.Cancel: %w", err)
	}

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("repo.Update: %w", err)
	}

	return order, nil
}
