package handler

import (
	"encoding/json"
	"net/http"

	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/utils"
	"github.com/go-chi/chi/v5"

	inventory "github.com/f4ke-n0name/autoparts-hub/gen/inventory"
)

type InventoryHandler struct {
	client inventory.InventoryServiceClient
}

func NewInventoryHandler(client inventory.InventoryServiceClient) *InventoryHandler {
	return &InventoryHandler{client: client}
}

// GetStock godoc
// @Summary      Get stock by part ID
// @Tags         inventory
// @Produce      json
// @Param        part_id path string true "Part ID"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /inventory/stock/{part_id} [get]
func (h *InventoryHandler) GetStock(w http.ResponseWriter, r *http.Request) {
	partID := chi.URLParam(r, "part_id")

	resp, err := h.client.GetStock(r.Context(), &inventory.GetStockRequest{PartId: partID})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// UpdateStock godoc
// @Summary      Update stock quantity
// @Tags         inventory
// @Accept       json
// @Produce      json
// @Param        request body UpdateStockRequest true "Update stock"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /inventory/stock [put]
func (h *InventoryHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var req UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.UpdateStock(r.Context(), &inventory.UpdateStockRequest{
		PartId:   req.PartID,
		Quantity: req.Quantity,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

type UpdateStockRequest struct {
	PartID   string `json:"part_id"`
	Quantity int32  `json:"quantity"`
}
