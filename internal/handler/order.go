package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/arthur-stone/order-service/internal/domain"
	"github.com/arthur-stone/order-service/internal/service"
)

type OrderHandler struct {
	svc service.OrderService
}

func NewOrderHandler(svc service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type CreateOrderRequest struct {
	CustomerID string             `json:"customer_id"`
	Items      []domain.OrderItem `json:"items"`
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	order, err := h.svc.CreateOrder(r.Context(), req.CustomerID, req.Items)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmptyCustomerID),
			errors.Is(err, domain.ErrEmptyItems),
			errors.Is(err, domain.ErrInvalidQuantity),
			errors.Is(err, domain.ErrInvalidPrice):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing order id")
		return
	}

	order, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			respondError(w, http.StatusNotFound, "order not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	orders, err := h.svc.ListOrders(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"items":  orders,
		"count":  len(orders),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing order id")
		return
	}

	order, err := h.svc.PayOrder(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			respondError(w, http.StatusNotFound, "order not found")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "missing order id")
		return
	}

	order, err := h.svc.CancelOrder(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			respondError(w, http.StatusNotFound, "order not found")
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, order)
}
