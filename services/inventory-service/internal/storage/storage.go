package storage

import (
	"context"

	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/domain/models"
)

type StockStorage interface {
	GetStock(ctx context.Context, partID string) (*models.Stock, error)
	UpdateStock(ctx context.Context, partID string, quantity int32) (*models.Stock, error)
	ReserveStock(ctx context.Context, orderID string, items []*models.ReserveItem) error
	SaveMovement(ctx context.Context, movement *models.StockMovement) error
	ReleaseStock(ctx context.Context, orderID string, items []*models.ReserveItem) error
	WriteOffStock(ctx context.Context, orderID string, items []*models.ReserveItem) error
}
