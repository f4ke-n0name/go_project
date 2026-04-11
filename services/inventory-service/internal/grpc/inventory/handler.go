package inventory

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	inventory "github.com/f4ke-n0name/autoparts-hub/gen/inventory"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/domain/models"
)

type InventoryService interface {
	GetStock(ctx context.Context, partID string) (*models.Stock, error)
	UpdateStock(ctx context.Context, partID string, quantity int32) (*models.Stock, error)
	ReserveStock(ctx context.Context, orderID string, items []*models.ReserveItem) error
}

type Handler struct {
	inventory.UnimplementedInventoryServiceServer
	log     *slog.Logger
	service InventoryService
}

func New(log *slog.Logger, service InventoryService) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) GetStock(ctx context.Context, req *inventory.GetStockRequest) (*inventory.GetStockResponse, error) {
	const op = "grpc.inventory.Handler.GetStock"
	log := h.log.With(slog.String("op", op))

	if req.GetPartId() == "" {
		return nil, status.Error(codes.InvalidArgument, "part_id is required")
	}

	stock, err := h.service.GetStock(ctx, req.GetPartId())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "stock not found")
		}
		log.Error("failed to get stock", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &inventory.GetStockResponse{Stock: toProtoStock(stock)}, nil
}

func (h *Handler) UpdateStock(ctx context.Context, req *inventory.UpdateStockRequest) (*inventory.UpdateStockResponse, error) {
	const op = "grpc.inventory.Handler.UpdateStock"
	log := h.log.With(slog.String("op", op))

	if req.GetPartId() == "" {
		return nil, status.Error(codes.InvalidArgument, "part_id is required")
	}
	if req.GetQuantity() == 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity cannot be zero")
	}

	stock, err := h.service.UpdateStock(ctx, req.GetPartId(), req.GetQuantity())
	if err != nil {
		log.Error("failed to update stock", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &inventory.UpdateStockResponse{Stock: toProtoStock(stock)}, nil
}

func (h *Handler) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	const op = "grpc.inventory.Handler.ReserveStock"
	log := h.log.With(slog.String("op", op), slog.String("order_id", req.GetOrderId()))

	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	items := make([]*models.ReserveItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, &models.ReserveItem{
			PartID:   item.GetPartId(),
			Quantity: item.GetQuantity(),
		})
	}

	if err := h.service.ReserveStock(ctx, req.GetOrderId(), items); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return &inventory.ReserveStockResponse{
				Success: false,
				Message: "part not found in stock",
			}, nil
		}
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return &inventory.ReserveStockResponse{
				Success: false,
				Message: "insufficient stock",
			}, nil
		}
		log.Error("failed to reserve stock", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &inventory.ReserveStockResponse{Success: true}, nil
}
func toProtoStock(s *models.Stock) *inventory.Stock {
	return &inventory.Stock{
		PartId:    s.PartID,
		Quantity:  s.Quantity,
		Reserved:  s.Reserved,
		Available: s.Available(),
	}
}
