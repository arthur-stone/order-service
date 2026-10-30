package domain

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrEmptyCustomerID         = errors.New("customer ID cannot be empty")
	ErrEmptyItems              = errors.New("order must contain at least one item")
	ErrInvalidQuantity         = errors.New("item quantity must be greater than zero")
	ErrInvalidPrice            = errors.New("item price must be positive")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

type OrderStatus string

const (
	StatusCreated  OrderStatus = "CREATED"
	StatusPaid     OrderStatus = "PAID"
	StatusCanceled OrderStatus = "CANCELED"
)

type OrderItem struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Title     string  `json:"title"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Order struct {
	ID          string      `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

func NewOrder(customerID string, items []OrderItem) (*Order, error) {
	if customerID == "" {
		return nil, ErrEmptyCustomerID
	}
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}

	var total float64
	for i := range items {
		if items[i].Quantity <= 0 {
			return nil, fmt.Errorf("%w: product %s", ErrInvalidQuantity, items[i].ProductID)
		}
		if items[i].Price <= 0 {
			return nil, fmt.Errorf("%w: product %s", ErrInvalidPrice, items[i].ProductID)
		}
		if items[i].ID == "" {
			items[i].ID = generateID("item")
		}
		total += float64(items[i].Quantity) * items[i].Price
	}

	now := time.Now().UTC()
	return &Order{
		ID:          generateID("ord"),
		CustomerID:  customerID,
		Items:       items,
		TotalAmount: total,
		Status:      StatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (o *Order) Pay() error {
	if o.Status != StatusCreated {
		return fmt.Errorf("%w: cannot pay order in status %s", ErrInvalidStatusTransition, o.Status)
	}
	o.Status = StatusPaid
	o.UpdatedAt = time.Now().UTC()
	return nil
}

func (o *Order) Cancel() error {
	if o.Status == StatusCanceled {
		return fmt.Errorf("%w: order is already canceled", ErrInvalidStatusTransition)
	}
	if o.Status == StatusPaid {
		return fmt.Errorf("%w: paid order cannot be canceled directly", ErrInvalidStatusTransition)
	}
	o.Status = StatusCanceled
	o.UpdatedAt = time.Now().UTC()
	return nil
}

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(b))
}
