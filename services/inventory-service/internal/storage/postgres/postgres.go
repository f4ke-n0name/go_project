package postgres

import (
	"context"
	"errors"
	"fmt"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Storage{db: pool}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func (s *Storage) GetStock(ctx context.Context, partID string) (*models.Stock, error) {
	query := `
		SELECT id::text, part_id::text, quantity, reserved, updated_at
		FROM stock WHERE part_id = $1`

	stock := &models.Stock{}
	err := s.db.QueryRow(ctx, query, partID).Scan(&stock.ID, &stock.PartID, &stock.Quantity, &stock.Reserved, &stock.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get stock: %w", err)
	}
	return stock, nil
}

func (s *Storage) UpdateStock(ctx context.Context, partID string, quantity int32) (*models.Stock, error) {
	query := `
		INSERT INTO stock (part_id, quantity)
		VALUES ($1, $2)
		ON CONFLICT (part_id) DO UPDATE
			SET quantity = stock.quantity + $2,
			    updated_at = NOW()
		RETURNING id::text, part_id::text, quantity, reserved, updated_at`

	stock := &models.Stock{}
	err := s.db.QueryRow(ctx, query, partID, quantity).Scan(&stock.ID, &stock.PartID, &stock.Quantity, &stock.Reserved, &stock.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update stock: %w", err)
	}
	return stock, nil
}

func (s *Storage) ReserveStock(ctx context.Context, orderID string, items []*models.ReserveItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		var stock models.Stock
		query := `
			SELECT id::text, quantity, reserved
			FROM stock
			WHERE part_id = $1
			FOR UPDATE`

		err = tx.QueryRow(ctx, query, item.PartID).Scan(&stock.ID, &stock.Quantity, &stock.Reserved)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("part %s not found in stock: %w", item.PartID, pkgerrors.ErrNotFound)
			}
			return fmt.Errorf("lock stock row: %w", err)
		}

		available := stock.Quantity - stock.Reserved
		if available < item.Quantity {
			return fmt.Errorf("insufficient stock for part %s: available %d, requested %d: %w",
				item.PartID, available, item.Quantity, pkgerrors.ErrInvalidInput)
		}

		queryStock := `
			UPDATE stock
			SET reserved = reserved + $1, updated_at = NOW()
			WHERE id = $2`

		_, err = tx.Exec(ctx, queryStock, item.Quantity, stock.ID)
		if err != nil {
			return fmt.Errorf("reserve stock: %w", err)
		}

		queryInsert := `
			INSERT INTO stock_movements (part_id, order_id, type, quantity)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(ctx, queryInsert, item.PartID, orderID, models.MovementTypeReserve, item.Quantity)
		if err != nil {
			return fmt.Errorf("save movement: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (s *Storage) SaveMovement(ctx context.Context, movement *models.StockMovement) error {
	query := `
		INSERT INTO stock_movements (part_id, order_id, type, quantity)
		VALUES ($1, $2, $3, $4)`

	_, err := s.db.Exec(ctx, query, movement.PartID, movement.OrderID, movement.Type, movement.Quantity)
	if err != nil {
		return fmt.Errorf("save movement: %w", err)
	}
	return nil
}

func (s *Storage) ReleaseStock(ctx context.Context, orderID string, items []*models.ReserveItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		var stockID string
		query := `
			SELECT id::text FROM stock
			WHERE part_id = $1
			FOR UPDATE`
		err = tx.QueryRow(ctx, query, item.PartID).Scan(&stockID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("part %s not found: %w", item.PartID, pkgerrors.ErrNotFound)
			}
			return fmt.Errorf("lock stock row: %w", err)
		}

		queryStock := `
			UPDATE stock
			SET reserved = GREATEST(reserved - $1, 0), updated_at = NOW()
			WHERE id = $2`
		_, err = tx.Exec(ctx, queryStock, item.Quantity, stockID)
		if err != nil {
			return fmt.Errorf("release stock: %w", err)
		}

		queryLog := `
			INSERT INTO stock_movements (part_id, order_id, type, quantity)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(ctx, queryLog, item.PartID, orderID, models.MovementTypeRelease, item.Quantity)
		if err != nil {
			return fmt.Errorf("save movement: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *Storage) WriteOffStock(ctx context.Context, orderID string, items []*models.ReserveItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		var stockID string
		query := `
			SELECT id::text FROM stock
			WHERE part_id = $1
			FOR UPDATE`
		err = tx.QueryRow(ctx, query, item.PartID).Scan(&stockID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("part %s not found: %w", item.PartID, pkgerrors.ErrNotFound)
			}
			return fmt.Errorf("lock stock row: %w", err)
		}
		queryStock := `
			UPDATE stock
			SET reserved = GREATEST(reserved - $1, 0),
			    quantity  = GREATEST(quantity - $1, 0),
			    updated_at = NOW()
			WHERE id = $2`
		_, err = tx.Exec(ctx, queryStock, item.Quantity, stockID)
		if err != nil {
			return fmt.Errorf("write off stock: %w", err)
		}

		queryLog := `
			INSERT INTO stock_movements (part_id, order_id, type, quantity)
			VALUES ($1, $2, $3, $4)`
		_, err = tx.Exec(ctx, queryLog, item.PartID, orderID, models.MovementTypeWriteoff, item.Quantity)
		if err != nil {
			return fmt.Errorf("save movement: %w", err)
		}
	}

	return tx.Commit(ctx)
}
