package storage

import (
	"context"

	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/domain/models"
)

type AutoPartStorage interface {
	CreatePart(ctx context.Context, autoPart *models.AutoPart) error
	GetPartByID(ctx context.Context, id string) (*models.AutoPart, error)
	SearchParts(ctx context.Context, filter SearchFilter) ([]*models.AutoPart, int32, error)
}

type CategoryStorage interface {
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type CompatibilityStorage interface {
	GetCompatibilityByPartID(ctx context.Context, id string) ([]*models.Compatibility, error)
}

type SearchFilter struct {
	Query      string
	CategoryID string
	Brand      string
	Make       string
	Model      string
	Year       int32
	Page       int32
	PageSize   int32
}
