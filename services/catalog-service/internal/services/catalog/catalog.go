package catalog

import (
	"context"
	"fmt"
	"log/slog"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/storage"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	log           *slog.Logger
	parts         storage.AutoPartStorage
	categories    storage.CategoryStorage
	compatibility storage.CompatibilityStorage
}

func New(
	log *slog.Logger,
	parts storage.AutoPartStorage,
	categories storage.CategoryStorage,
	compatibility storage.CompatibilityStorage,
) *Service {
	return &Service{
		log:           log,
		parts:         parts,
		categories:    categories,
		compatibility: compatibility,
	}
}

func (s *Service) CreatePart(ctx context.Context, part *models.AutoPart) (*models.AutoPart, error) {
	const op = "catalog.Service.CreatePart"
	log := s.log.With(slog.String("op", op))

	log.Info("creating part", slog.String("sku", part.SKU))

	if err := s.parts.CreatePart(ctx, part); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
			log.Warn("part already exists", slog.String("sku", part.SKU))
			return nil, pkgerrors.ErrAlreadyExists
		}
		log.Error("failed to create part", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return part, nil
}

func (s *Service) GetPart(ctx context.Context, id string) (*models.AutoPart, []*models.Compatibility, error) {
	const op = "catalog.Service.GetPart"
	log := s.log.With(slog.String("op", op), slog.String("id", id))

	log.Info("getting part")

	var part *models.AutoPart
	var compatibility []*models.Compatibility

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		part, err = s.parts.GetPartByID(gCtx, id)
		if err != nil {
			return fmt.Errorf("get part: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		var err error
		compatibility, err = s.compatibility.GetCompatibilityByPartID(gCtx, id)
		if err != nil {
			return fmt.Errorf("get compatibility: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			log.Warn("part not found")
			return nil, nil, pkgerrors.ErrNotFound
		}
		log.Error("failed to get part", logger.Err(err))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return part, compatibility, nil
}

func (s *Service) SearchParts(ctx context.Context, filter storage.SearchFilter) ([]*models.AutoPart, int32, error) {
	const op = "catalog.Service.SearchParts"
	log := s.log.With(slog.String("op", op))

	log.Info("searching parts",
		slog.String("query", filter.Query),
		slog.String("brand", filter.Brand),
		slog.String("make", filter.Make),
		slog.String("model", filter.Model),
	)

	parts, total, err := s.parts.SearchParts(ctx, filter)
	if err != nil {
		log.Error("failed to search parts", logger.Err(err))
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return parts, total, nil
}

func (s *Service) ListCategories(ctx context.Context) ([]*models.Category, error) {
	const op = "catalog.Service.ListCategories"

	categories, err := s.categories.ListCategories(ctx)
	if err != nil {
		s.log.Error("failed to list categories", slog.String("op", op), logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return categories, nil
}
