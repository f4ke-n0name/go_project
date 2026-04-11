package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/storage"
)

type Service struct {
	log   *slog.Logger
	stock storage.StockStorage
}

func New(log *slog.Logger, stock storage.StockStorage) *Service {
	return &Service{log: log, stock: stock}
}

func (s *Service) GetStock(ctx context.Context, partID string) (*models.Stock, error) {
	const op = "inventory.Service.GetStock"

	stock, err := s.stock.GetStock(ctx, partID)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		s.log.Error("failed to get stock", slog.String("op", op), logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stock, nil
}

func (s *Service) UpdateStock(ctx context.Context, partID string, quantity int32) (*models.Stock, error) {
	const op = "inventory.Service.UpdateStock"
	log := s.log.With(slog.String("op", op), slog.String("part_id", partID))

	log.Info("updating stock", slog.Int("quantity", int(quantity)))

	stock, err := s.stock.UpdateStock(ctx, partID, quantity)
	if err != nil {
		log.Error("failed to update stock", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stock, nil
}

func (s *Service) ReserveStock(ctx context.Context, orderID string, items []*models.ReserveItem) error {
	const op = "inventory.Service.ReserveStock"
	log := s.log.With(slog.String("op", op), slog.String("order_id", orderID))

	log.Info("reserving stock", slog.Int("items_count", len(items)))

	if err := s.stock.ReserveStock(ctx, orderID, items); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return pkgerrors.ErrNotFound
		}
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			return pkgerrors.ErrInvalidInput
		}
		log.Error("failed to reserve stock", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("stock reserved successfully")
	return nil
}
func (s *Service) HandleOrderCreated(ctx context.Context, payload []byte) error {
	const op = "inventory.Service.HandleOrderCreated"
	log := s.log.With(slog.String("op", op))

	var event struct {
		OrderID string `json:"order_id"`
		Items   []struct {
			PartID   string  `json:"part_id"`
			Quantity int32   `json:"quantity"`
			Price    float64 `json:"price"`
		} `json:"items"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log = log.With(slog.String("order_id", event.OrderID))
	log.Info("handling order.created")

	items := make([]*models.ReserveItem, 0, len(event.Items))
	for _, item := range event.Items {
		items = append(items, &models.ReserveItem{
			PartID:   item.PartID,
			Quantity: item.Quantity,
		})
	}

	if err := s.ReserveStock(ctx, event.OrderID, items); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrInvalidInput) {
			log.Warn("insufficient stock", slog.String("order_id", event.OrderID))
			// TODO: publish inventory.insufficient
			return nil // не ретраим — это бизнес ошибка, не техническая
		}
		return fmt.Errorf("reserve stock: %w", err)
	}

	log.Info("stock reserved for order")
	// TODO: publish inventory.reserved
	return nil
}

func (s *Service) HandleOrderStatusChanged(ctx context.Context, payload []byte) error {
	const op = "inventory.Service.HandleOrderStatusChanged"
	log := s.log.With(slog.String("op", op))

	var event struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
		Items   []struct {
			PartID   string `json:"part_id"`
			Quantity int32  `json:"quantity"`
		} `json:"items"`
	}

	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	log = log.With(slog.String("order_id", event.OrderID), slog.String("status", event.Status))
	log.Info("handling order.status_changed")

	items := make([]*models.ReserveItem, 0, len(event.Items))
	for _, item := range event.Items {
		items = append(items, &models.ReserveItem{
			PartID:   item.PartID,
			Quantity: item.Quantity,
		})
	}

	switch event.Status {
	case "CANCELLED":
		if err := s.stock.ReleaseStock(ctx, event.OrderID, items); err != nil {
			log.Error("failed to release stock", logger.Err(err))
			return fmt.Errorf("release stock: %w", err)
		}
		log.Info("stock released for cancelled order")

	case "DELIVERED":
		if err := s.stock.WriteOffStock(ctx, event.OrderID, items); err != nil {
			log.Error("failed to write off stock", logger.Err(err))
			return fmt.Errorf("write off stock: %w", err)
		}
		log.Info("stock written off for delivered order")

	default:
		log.Debug("no stock action for status", slog.String("status", event.Status))
	}

	return nil
}
