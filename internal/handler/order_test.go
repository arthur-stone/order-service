package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arthur-stone/order-service/internal/domain"
	"github.com/arthur-stone/order-service/internal/handler"
	"github.com/arthur-stone/order-service/internal/repository"
	"github.com/arthur-stone/order-service/internal/service"
)

func setupTestRouter() http.Handler {
	repo := repository.NewMemoryOrderRepository()
	svc := service.NewOrderService(repo)
	h := handler.NewOrderHandler(svc)
	hh := handler.NewHealthHandler(handler.BuildInfo{
		Version:   "1.0.0-test",
		Commit:    "abcdef",
		BuildDate: "2026-10-30",
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", hh.Healthz)
	mux.HandleFunc("GET /ready", hh.Ready)
	mux.HandleFunc("GET /version", hh.Version)

	mux.HandleFunc("POST /api/v1/orders", h.Create)
	mux.HandleFunc("GET /api/v1/orders/{id}", h.GetByID)
	mux.HandleFunc("GET /api/v1/orders", h.List)
	mux.HandleFunc("POST /api/v1/orders/{id}/pay", h.Pay)
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", h.Cancel)

	return mux
}

func TestHealthEndpoints(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var res map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if res["status"] != "UP" {
		t.Errorf("expected UP, got %s", res["status"])
	}
}

func TestOrderHandler_CreateAndGet(t *testing.T) {
	router := setupTestRouter()

	body := handler.CreateOrderRequest{
		CustomerID: "cust_abc",
		Items: []domain.OrderItem{
			{ProductID: "p1", Title: "Phone", Quantity: 1, Price: 500},
		},
	}
	data, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(data))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", rr.Code, rr.Body.String())
	}

	var created domain.Order
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Get by ID
	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+created.ID, nil)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getRR.Code)
	}
}
