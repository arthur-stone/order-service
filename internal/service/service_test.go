package service_test

import (
	"context"
	"testing"

	"github.com/arthur-stone/order-service/internal/domain"
	"github.com/arthur-stone/order-service/internal/repository"
	"github.com/arthur-stone/order-service/internal/service"
)

func TestOrderService_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewMemoryOrderRepository()
	svc := service.NewOrderService(repo)

	// Create
	items := []domain.OrderItem{
		{ProductID: "p-1", Title: "Keyboard", Quantity: 1, Price: 80.0},
	}
	created, err := svc.CreateOrder(ctx, "cust-1", items)
	if err != nil {
		t.Fatalf("unexpected error creating order: %v", err)
	}

	// Get
	found, err := svc.GetOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error getting order: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, found.ID)
	}

	// List
	orders, err := svc.ListOrders(ctx, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error listing orders: %v", err)
	}
	if len(orders) != 1 {
		t.Errorf("expected 1 order, got %d", len(orders))
	}

	// Pay
	paid, err := svc.PayOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected error paying order: %v", err)
	}
	if paid.Status != domain.StatusPaid {
		t.Errorf("expected status PAID, got %s", paid.Status)
	}

	// Cancel paid should fail
	_, err = svc.CancelOrder(ctx, created.ID)
	if err == nil {
		t.Errorf("expected error canceling paid order")
	}
}
