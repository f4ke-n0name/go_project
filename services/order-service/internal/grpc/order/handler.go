package order

import (
	"context"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	order "github.com/f4ke-n0name/autoparts-hub/gen/order"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/domain/models"
	orderstorage "github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/storage"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/utils"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, items []*models.OrderItem) (*models.Order, error)
	GetOrder(ctx context.Context, id string) (*models.Order, error)
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus) (*models.Order, error)
	ListOrders(ctx context.Context, filter orderstorage.ListOrdersFilter) ([]*models.Order, int32, error)
}

type Handler struct {
	order.UnimplementedOrderServiceServer
	log     *slog.Logger
	service OrderService
}

func New(log *slog.Logger, service OrderService) *Handler {
	return &Handler{log: log, service: service}
}

func (h *Handler) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.CreateOrderResponse, error) {
	const op = "grpc.order.Handler.CreateOrder"
	log := h.log.With(slog.String("op", op))

	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}

	items := make([]*models.OrderItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, &models.OrderItem{
			PartID:   item.GetPartId(),
			Quantity: item.GetQuantity(),
			Price:    item.GetPrice(),
		})
	}

	newOrder, err := h.service.CreateOrder(ctx, req.GetUserId(), items)
	if err != nil {
		log.Error("failed to create order", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &order.CreateOrderResponse{Order: utils.ToProtoOrder(newOrder)}, nil
}

func (h *Handler) GetOrder(ctx context.Context, req *order.GetOrderRequest) (*order.GetOrderResponse, error) {
	const op = "grpc.order.Handler.GetOrder"
	log := h.log.With(slog.String("op", op), slog.String("id", req.GetId()))

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	gettedOrder, err := h.service.GetOrder(ctx, req.GetId())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		log.Error("failed to get order", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &order.GetOrderResponse{Order: utils.ToProtoOrder(gettedOrder)}, nil
}

func (h *Handler) UpdateStatus(ctx context.Context, req *order.UpdateStatusRequest) (*order.UpdateStatusResponse, error) {
	const op = "grpc.order.Handler.UpdateStatus"
	log := h.log.With(slog.String("op", op), slog.String("id", req.GetId()))
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.GetStatus() == order.OrderStatus_ORDER_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	updatedOrder, err := h.service.UpdateStatus(ctx, req.GetId(), utils.ProtoStatusToModel(req.GetStatus()))
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		log.Error("failed to update status", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &order.UpdateStatusResponse{Order: utils.ToProtoOrder(updatedOrder)}, nil
}

func (h *Handler) ListOrders(ctx context.Context, req *order.ListOrdersRequest) (*order.ListOrdersResponse, error) {
	const op = "grpc.order.Handler.ListOrders"
	log := h.log.With(slog.String("op", op))

	orders, total, err := h.service.ListOrders(ctx, orderstorage.ListOrdersFilter{
		UserID:   req.GetUserId(),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		log.Error("failed to list orders", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoOrders := make([]*order.Order, 0, len(orders))
	for _, o := range orders {
		protoOrders = append(protoOrders, utils.ToProtoOrder(o))
	}

	return &order.ListOrdersResponse{Orders: protoOrders, Total: total}, nil
}
