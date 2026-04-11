package storage

import (
	"context"

	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/domain/models"
)

type ListOrdersFilter struct {
	UserID   string
	Page     int32
	PageSize int32
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error)
	ListOrders(ctx context.Context, filter ListOrdersFilter) ([]*models.Order, int32, error)
}

type OutboxStorage interface {
	SaveEvent(ctx context.Context, topic string, payload []byte) error
	GetUnsent(ctx context.Context, limit int32) ([]*models.OutboxEvent, error)
	MarkSent(ctx context.Context, id string) error
}
