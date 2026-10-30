package domain_test

import (
	"testing"

	"github.com/arthur-stone/order-service/internal/domain"
)

func TestNewOrder_Success(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "prod_1", Title: "Laptop", Quantity: 1, Price: 1200.0},
		{ProductID: "prod_2", Title: "Mouse", Quantity: 2, Price: 25.5},
	}

	order, err := domain.NewOrder("cust_123", items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.ID == "" {
		t.Errorf("expected non-empty ID")
	}
	if order.CustomerID != "cust_123" {
		t.Errorf("expected customer ID 'cust_123', got %s", order.CustomerID)
	}
	if order.Status != domain.StatusCreated {
		t.Errorf("expected status 'CREATED', got %s", order.Status)
	}
	expectedTotal := 1200.0 + 2*25.5
	if order.TotalAmount != expectedTotal {
		t.Errorf("expected total amount %f, got %f", expectedTotal, order.TotalAmount)
	}
}

func TestNewOrder_Validations(t *testing.T) {
	t.Run("empty customer ID", func(t *testing.T) {
		_, err := domain.NewOrder("", []domain.OrderItem{{ProductID: "p1", Quantity: 1, Price: 10}})
		if err != domain.ErrEmptyCustomerID {
			t.Errorf("expected ErrEmptyCustomerID, got %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := domain.NewOrder("c1", nil)
		if err != domain.ErrEmptyItems {
			t.Errorf("expected ErrEmptyItems, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := domain.NewOrder("c1", []domain.OrderItem{{ProductID: "p1", Quantity: 0, Price: 10}})
		if err == nil {
			t.Fatalf("expected error for zero quantity")
		}
	})

	t.Run("invalid price", func(t *testing.T) {
		_, err := domain.NewOrder("c1", []domain.OrderItem{{ProductID: "p1", Quantity: 1, Price: -5}})
		if err == nil {
			t.Fatalf("expected error for negative price")
		}
	})
}

func TestOrder_StatusTransitions(t *testing.T) {
	items := []domain.OrderItem{
		{ProductID: "p1", Quantity: 1, Price: 100},
	}
	order, err := domain.NewOrder("cust_1", items)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Pay
	if err := order.Pay(); err != nil {
		t.Fatalf("failed to pay: %v", err)
	}
	if order.Status != domain.StatusPaid {
		t.Errorf("expected status PAID, got %s", order.Status)
	}

	// Cannot pay again
	if err := order.Pay(); err == nil {
		t.Errorf("expected error when paying already paid order")
	}

	// Cannot cancel paid order directly
	if err := order.Cancel(); err == nil {
		t.Errorf("expected error when cancelling paid order")
	}

	// Test fresh order cancel
	freshOrder, _ := domain.NewOrder("cust_2", items)
	if err := freshOrder.Cancel(); err != nil {
		t.Fatalf("failed to cancel fresh order: %v", err)
	}
	if freshOrder.Status != domain.StatusCancelled {
		t.Errorf("expected status CANCELLED, got %s", freshOrder.Status)
	}
}
