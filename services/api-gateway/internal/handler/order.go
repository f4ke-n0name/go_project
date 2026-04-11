package handler

import (
	"encoding/json"
	"net/http"

	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/utils"
	"github.com/go-chi/chi/v5"

	order "github.com/f4ke-n0name/autoparts-hub/gen/order"
	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/middleware"
)

type OrderHandler struct {
	client order.OrderServiceClient
}

func NewOrderHandler(client order.OrderServiceClient) *OrderHandler {
	return &OrderHandler{client: client}
}

// CreateOrder godoc
// @Summary      Create order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body CreateOrderRequest true "Create order request"
// @Success      201 {object} map[string]any
// @Security     BearerAuth
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items := make([]*order.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &order.OrderItem{
			PartId:   item.PartID,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	resp, err := h.client.CreateOrder(r.Context(), &order.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// GetOrder godoc
// @Summary      Get order by ID
// @Tags         orders
// @Produce      json
// @Param        id path string true "Order ID"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resp, err := h.client.GetOrder(r.Context(), &order.GetOrderRequest{Id: id})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// ListOrders godoc
// @Summary      List orders
// @Tags         orders
// @Produce      json
// @Param        page      query int false "Page"
// @Param        page_size query int false "Page size"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID := middleware.GetUserID(r.Context())

	resp, err := h.client.ListOrders(r.Context(), &order.ListOrdersRequest{
		UserId:   userID,
		Page:     int32(utils.ParseInt(q.Get("page"))),
		PageSize: int32(utils.ParseInt(q.Get("page_size"))),
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// UpdateStatus godoc
// @Summary      Update order status
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id      path string            true "Order ID"
// @Param        request body UpdateStatusRequest true "Status update"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateStatus(r.Context(), &order.UpdateStatusRequest{
		Id:     id,
		Status: order.OrderStatus(req.Status),
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	PartID   string  `json:"part_id"`
	Quantity int32   `json:"quantity"`
	Price    float64 `json:"price"`
}

type UpdateStatusRequest struct {
	Status int32 `json:"status"`
}
