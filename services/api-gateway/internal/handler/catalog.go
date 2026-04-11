package handler

import (
	"encoding/json"
	"net/http"

	"github.com/f4ke-n0name/autoparts-hub/services/api-gateway/internal/utils"
	"github.com/go-chi/chi/v5"

	catalog "github.com/f4ke-n0name/autoparts-hub/gen/catalog"
)

type CatalogHandler struct {
	client catalog.CatalogServiceClient
}

func NewCatalogHandler(client catalog.CatalogServiceClient) *CatalogHandler {
	return &CatalogHandler{client: client}
}

// SearchParts godoc
// @Summary      Search parts
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        query     query string false "Search query"
// @Param        brand     query string false "Brand filter"
// @Param        make      query string false "Car make"
// @Param        model     query string false "Car model"
// @Param        year      query int    false "Car year"
// @Param        page      query int    false "Page number"
// @Param        page_size query int    false "Page size"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /catalog/parts [get]
func (h *CatalogHandler) SearchParts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	resp, err := h.client.SearchParts(r.Context(), &catalog.SearchPartsRequest{
		Query:    q.Get("query"),
		Brand:    q.Get("brand"),
		Make:     q.Get("make"),
		Model:    q.Get("model"),
		Year:     int32(utils.ParseInt(q.Get("year"))),
		Page:     int32(utils.ParseInt(q.Get("page"))),
		PageSize: int32(utils.ParseInt(q.Get("page_size"))),
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// GetPart godoc
// @Summary      Get part by ID
// @Tags         catalog
// @Produce      json
// @Param        id path string true "Part ID"
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /catalog/parts/{id} [get]
func (h *CatalogHandler) GetPart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resp, err := h.client.GetPart(r.Context(), &catalog.GetPartRequest{Id: id})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// CreatePart godoc
// @Summary      Create part
// @Tags         catalog
// @Accept       json
// @Produce      json
// @Param        request body CreatePartRequest true "Create part request"
// @Success      201 {object} map[string]any
// @Security     BearerAuth
// @Router       /catalog/parts [post]
func (h *CatalogHandler) CreatePart(w http.ResponseWriter, r *http.Request) {
	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.client.CreatePart(r.Context(), &catalog.CreateRequest{
		Sku:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		Brand:       req.Brand,
		Price:       req.Price,
		CategoryId:  req.CategoryID,
	})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, resp)
}

// ListCategories godoc
// @Summary      List categories
// @Tags         catalog
// @Produce      json
// @Success      200 {object} map[string]any
// @Security     BearerAuth
// @Router       /catalog/categories [get]
func (h *CatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.ListCategories(r.Context(), &catalog.ListCategoriesRequest{})
	if err != nil {
		utils.WriteGRPCError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

type CreatePartRequest struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Brand       string  `json:"brand"`
	Price       float64 `json:"price"`
	CategoryID  string  `json:"category_id"`
}
