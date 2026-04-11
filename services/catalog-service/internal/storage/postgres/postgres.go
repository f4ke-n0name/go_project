package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &Storage{db: pool}, nil
}

func (s *Storage) CreatePart(ctx context.Context, autoPart *models.AutoPart) error {
	query := `
		INSERT INTO parts(sku, name, description, brand, price, category_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, stock, created_at, updated_at`

	err := s.db.QueryRow(
		ctx,
		query,
		autoPart.SKU,
		autoPart.Name,
		autoPart.Description,
		autoPart.Brand,
		autoPart.Price,
		nullableString(autoPart.CategoryID)).Scan(&autoPart.ID, &autoPart.Stock, &autoPart.CreatedAt, &autoPart.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return pkgerrors.ErrAlreadyExists
		}
		return fmt.Errorf("get auto part by id : %w", err)
	}
	return nil
}

func (s *Storage) GetPartByID(ctx context.Context, id string) (*models.AutoPart, error) {
	query := `
		SELECT id, sku, name, description, brand, price, COALESCE(category_id::text, '') as category_id, stock, created_at, updated_at
		FROM parts
		WHERE id = $1`

	part := &models.AutoPart{}
	err := s.db.QueryRow(ctx, query, id).Scan(
		&part.ID,
		&part.SKU,
		&part.Name,
		&part.Description,
		&part.Brand,
		&part.Price,
		&part.CategoryID,
		&part.Stock,
		&part.CreatedAt,
		&part.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get auto part by id : %w", err)
	}
	return part, nil
}

func (s *Storage) SearchParts(ctx context.Context, filter storage.SearchFilter) ([]*models.AutoPart, int32, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	q := psql.Select(
		"id", "sku", "name", "description", "brand", "price",
		"COALESCE(category_id::text, '') as category_id",
		"stock", "created_at", "updated_at",
	).From("parts")

	if filter.Query != "" {
		q = q.Where("to_tsvector('english', name || ' ' || brand) @@ plainto_tsquery('english', ?)", filter.Query)
	}
	if filter.Brand != "" {
		q = q.Where("brand ILIKE ?", "%"+filter.Brand+"%")
	}
	if filter.CategoryID != "" {
		q = q.Where(sq.Eq{"category_id": filter.CategoryID})
	}
	if filter.Make != "" {
		q = q.Where("id IN (SELECT part_id FROM compatibility WHERE make ILIKE ?)", "%"+filter.Make+"%")
	}
	if filter.Model != "" {
		q = q.Where("id IN (SELECT part_id FROM compatibility WHERE model ILIKE ?)", "%"+filter.Model+"%")
	}
	if filter.Year > 0 {
		q = q.Where("id IN (SELECT part_id FROM compatibility WHERE year_from <= ? AND year_to >= ?)", filter.Year, filter.Year)
	}

	countSQL, args, err := psql.Select("COUNT(*)").
		FromSelect(q, "sub").
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}

	var total int32
	if err = s.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count parts: %w", err)
	}

	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	q = q.OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build search query: %w", err)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("search parts: %w", err)
	}
	defer rows.Close()

	var parts []*models.AutoPart
	for rows.Next() {
		part := &models.AutoPart{}
		if err = rows.Scan(
			&part.ID, &part.SKU, &part.Name, &part.Description,
			&part.Brand, &part.Price, &part.CategoryID,
			&part.Stock, &part.CreatedAt, &part.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan part: %w", err)
		}
		parts = append(parts, part)
	}

	return parts, total, nil
}

func (s *Storage) ListCategories(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, COALESCE(parent_id::text, ''), created_at
    	FROM categories
    	ORDER BY name`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		if err := rows.Scan(&category.ID, &category.Name, &category.ParentID, &category.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (s *Storage) GetCompatibilityByPartID(ctx context.Context, id string) ([]*models.Compatibility, error) {
	query := `
			SELECT part_id, make, model, year_from, year_to
			FROM compatibility
			WHERE id = $1`

	rows, err := s.db.Query(ctx, query, id)

	if err != nil {
		return nil, fmt.Errorf("get compatibility by part_id: %w", err)
	}
	defer rows.Close()
	var items []*models.Compatibility
	for rows.Next() {
		compatibility := &models.Compatibility{}
		if err = rows.Scan(&compatibility.PartID, &compatibility.Make, &compatibility.Model, &compatibility.YearFrom); err != nil {
			return nil, fmt.Errorf("scan compatibility: %w", err)
		}
		items = append(items, compatibility)
	}
	return items, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
