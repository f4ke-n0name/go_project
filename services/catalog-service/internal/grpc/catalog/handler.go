package catalog

import (
	"context"
	"log/slog"

	catalog "github.com/f4ke-n0name/autoparts-hub/gen/catalog"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService interface {
	CreatePart(ctx context.Context, part *models.AutoPart) (*models.AutoPart, error)
	GetPart(ctx context.Context, id string) (*models.AutoPart, []*models.Compatibility, error)
	SearchParts(ctx context.Context, filter storage.SearchFilter) ([]*models.AutoPart, int32, error)
	ListCategories(ctx context.Context) ([]*models.Category, error)
}

type Handler struct {
	catalog.UnimplementedCatalogServiceServer
	log     *slog.Logger
	service CatalogService
}

func New(log *slog.Logger, catalogService CatalogService) *Handler {
	return &Handler{
		log:     log,
		service: catalogService,
	}
}

func (h *Handler) CreatePart(ctx context.Context, req *catalog.CreateRequest) (*catalog.CreateResponse, error) {
	const op = "grpc.catalog.Handler.CreatePart"
	log := h.log.With(slog.String("op", op))

	if req.GetName() == "" || req.GetSku() == "" || req.GetBrand() == "" {
		return nil, status.Error(codes.InvalidArgument, "name, sku and brand are required")
	}

	part, err := h.service.CreatePart(ctx, &models.AutoPart{
		SKU:         req.GetSku(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Brand:       req.GetBrand(),
		Price:       req.GetPrice(),
		CategoryID:  req.GetCategoryId(),
	})

	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "part with SKU already exists")
		}
		log.Error("failed to create part", logger.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &catalog.CreateResponse{Part: toProtoPart(part)}, nil
}

func (h *Handler) GetPart(ctx context.Context, req *catalog.GetPartRequest) (*catalog.GetPartResponse, error) {
	const op = "grpc.catalog.Handler.GetPart"
	log := h.log.With(slog.String("op", op))

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	part, compatibility, err := h.service.GetPart(ctx, req.GetId())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "part not found")
		}
		log.Error("failed to get part", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoCompat := make([]*catalog.Compatibility, 0, len(compatibility))
	for _, comp := range compatibility {
		protoCompat = append(protoCompat, &catalog.Compatibility{
			Make:     comp.Make,
			Model:    comp.Model,
			YearFrom: comp.YearFrom,
			YearTo:   comp.YearTo,
		})
	}

	return &catalog.GetPartResponse{Part: toProtoPart(part), Compatibility: protoCompat}, nil
}

func (h *Handler) SearchParts(ctx context.Context, req *catalog.SearchPartsRequest) (*catalog.SearchPartsResponse, error) {
	const op = "grpc.catalog.Handler.SearchParts"
	log := h.log.With(slog.String("op", op))

	parts, total, err := h.service.SearchParts(ctx, storage.SearchFilter{
		Query:      req.GetQuery(),
		CategoryID: req.GetCategoryId(),
		Brand:      req.GetBrand(),
		Make:       req.GetMake(),
		Model:      req.GetModel(),
		Year:       req.GetYear(),
		Page:       req.GetPage(),
		PageSize:   req.GetPageSize(),
	})
	if err != nil {
		log.Error("failed to search parts", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoParts := make([]*catalog.AutoPart, 0, len(parts))
	for _, part := range parts {
		protoParts = append(protoParts, toProtoPart(part))
	}

	return &catalog.SearchPartsResponse{Parts: protoParts, Total: total}, nil
}

func (h *Handler) ListCategories(ctx context.Context, _ *catalog.ListCategoriesRequest) (*catalog.ListCategoriesResponse, error) {
	const op = "grpc.catalog.Handler.ListCategories"
	log := h.log.With(slog.String("op", op))

	categories, err := h.service.ListCategories(ctx)
	if err != nil {
		log.Error("failed to list categories", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	protoCategories := make([]*catalog.Category, 0, len(categories))
	for _, c := range categories {
		protoCategories = append(protoCategories, &catalog.Category{
			Id:       c.ID,
			Name:     c.Name,
			ParentId: c.ParentID,
		})
	}

	return &catalog.ListCategoriesResponse{Categories: protoCategories}, nil
}
func toProtoPart(part *models.AutoPart) *catalog.AutoPart {
	return &catalog.AutoPart{
		Id:          part.ID,
		Sku:         part.SKU,
		Name:        part.Name,
		Description: part.Description,
		Brand:       part.Brand,
		Price:       part.Price,
		CategoryId:  part.CategoryID,
		Stock:       part.Stock,
	}
}
