package postgres

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/storage"
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

func (s *Storage) CreateOrder(ctx context.Context, order *models.Order) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO orders (user_id, status, total_price)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(ctx, query, order.UserID, order.Status, order.TotalPrice).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	for _, item := range order.Items {

		queryItem := `
				INSERT INTO order_items (order_id, part_id, quantity, price)
				VALUES ($1, $2, $3, $4)
				RETURNING id`
		err := tx.QueryRow(ctx, queryItem, order.ID, item.PartID, item.Quantity, item.Price).Scan(&item.ID)
		if err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
		item.OrderID = order.ID
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
func (s *Storage) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	query := `
		SELECT id::text, user_id::text, status, total_price, created_at, updated_at
		FROM orders WHERE id = $1`

	order := &models.Order{}
	err := s.db.QueryRow(ctx, query, orderID).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}

	queryItem := `
			SELECT id::text, order_id::text, part_id::text, quantity, price
			FROM order_items WHERE order_id = $1`

	rows, err := s.db.Query(ctx, queryItem, orderID)

	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		item := &models.OrderItem{}
		if err = rows.Scan(&item.ID, &item.OrderID, &item.PartID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}
func (s *Storage) UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error) {
	query := `
		UPDATE orders SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, status, total_price, created_at, updated_at`

	order := &models.Order{}
	err := s.db.QueryRow(ctx, query, status, orderID).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("update status: %w", err)
	}

	return s.GetOrderByID(ctx, orderID)
}
func (s *Storage) ListOrders(ctx context.Context, filter storage.ListOrdersFilter) ([]*models.Order, int32, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	q := psql.Select("id", "user_id", "status", "total_price", "created_at", "updated_at").
		From("orders")

	if filter.UserID != "" {
		q = q.Where(sq.Eq{"user_id": filter.UserID})
	}

	countSQL, args, err := psql.Select("COUNT(*)").FromSelect(q, "sub").ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}

	var total int32
	if err = s.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	q = q.OrderBy("created_at DESC").
		Limit(uint64(pageSize)).
		Offset(uint64((page - 1) * pageSize))

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list query: %w", err)
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		o := &models.Order{}
		if err = rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}

	return orders, total, nil
}

func (s *Storage) SaveEvent(ctx context.Context, topic string, payload []byte) error {
	query := `
		INSERT INTO outbox (topic, payload)
		VALUES ($1, $2)`

	_, err := s.db.Exec(ctx, query, topic, payload)
	if err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}
	return nil
}
func (s *Storage) GetUnsent(ctx context.Context, limit int32) ([]*models.OutboxEvent, error) {
	query := `
		SELECT id, topic, payload, created_at
		FROM outbox
		WHERE sent_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get unsent events: %w", err)
	}
	defer rows.Close()

	var events []*models.OutboxEvent
	for rows.Next() {
		e := &models.OutboxEvent{}
		if err = rows.Scan(&e.ID, &e.Topic, &e.Payload, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, e)
	}

	return events, nil

}
func (s *Storage) MarkSent(ctx context.Context, id string) error {
	query := `UPDATE outbox SET sent_at = NOW() WHERE id = $1`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}
	return nil
}
