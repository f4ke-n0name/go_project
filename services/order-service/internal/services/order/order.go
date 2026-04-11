package order

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/storage"
)

const (
	topicOrderCreated       = "order.created"
	topicOrderStatusChanged = "order.status_changed"
)

type KafkaProducer interface {
	Produce(ctx context.Context, topic string, payload []byte) error
}

type Service struct {
	log      *slog.Logger
	orders   storage.OrderStorage
	outbox   storage.OutboxStorage
	producer KafkaProducer
}

func New(log *slog.Logger, orders storage.OrderStorage, outbox storage.OutboxStorage, producer KafkaProducer) *Service {
	return &Service{log: log, orders: orders, outbox: outbox, producer: producer}
}

func (s *Service) CreateOrder(ctx context.Context, userID string, items []*models.OrderItem) (*models.Order, error) {
	const op = "order.Service.CreateOrder"
	log := s.log.With(slog.String("op", op), slog.String("user_id", userID))
	log.Info("creating order")

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &models.Order{
		UserID:     userID,
		Status:     models.OrderStatusPending,
		TotalPrice: total,
		Items:      items,
	}

	if err := s.orders.CreateOrder(ctx, order); err != nil {
		log.Error("failed to create order", logger.Err(err))
		return nil, fmt.Errorf("%s: create order: %w", op, err)
	}

	payload, err := json.Marshal(map[string]any{
		"order_id": order.ID,
		"user_id":  userID,
		"total":    total,
		"items":    items,
		"created":  order.CreatedAt,
	})

	if err != nil {
		log.Error("failed to marshal outbox payload", logger.Err(err))
		return order, nil
	}

	if err = s.outbox.SaveEvent(ctx, topicOrderCreated, payload); err != nil {
		log.Error("failed to save outbox event", logger.Err(err))
		return order, nil
	}

	log.Info("created order")
	return order, nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (*models.Order, error) {
	const op = "order.Service.GetOrder"
	order, err := s.orders.GetOrderByID(ctx, id)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		s.log.Error("failed to get order", slog.String("op", op), logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, status models.OrderStatus) (*models.Order, error) {
	const op = "order.Service.UpdateStatus"
	log := s.log.With(slog.String("op", op), slog.String("order_id", id))
	log.Info("updating order status", slog.String("status", string(status)))

	order, err := s.orders.UpdateStatus(ctx, id, status)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return nil, pkgerrors.ErrNotFound
		}
		log.Error("failed to update status", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	payload, err := json.Marshal(map[string]any{
		"order_id":   id,
		"status":     status,
		"items":      order.Items,
		"updated_at": time.Now(),
	})
	if err == nil {
		if err = s.outbox.SaveEvent(ctx, topicOrderStatusChanged, payload); err != nil {
			log.Error("failed to save status changed event", logger.Err(err))
		}
	}

	return order, nil
}

func (s *Service) ListOrders(ctx context.Context, filter storage.ListOrdersFilter) ([]*models.Order, int32, error) {
	const op = "order.Service.ListOrders"
	orders, total, err := s.orders.ListOrders(ctx, filter)
	if err != nil {
		s.log.Error("failed to list orders", slog.String("op", op), logger.Err(err))
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return orders, total, nil
}

func (s *Service) StartOutboxWorker(ctx context.Context) {
	const op = "order.Service.OutboxWorker"
	log := s.log.With(slog.String("op", op))
	log.Info("starting outbox worker")

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("outbox worker stopped")
				return
			case <-ticker.C:
				s.processOutbox(ctx)
			}
		}
	}()
}

func (s *Service) processOutbox(ctx context.Context) {
	events, err := s.outbox.GetUnsent(ctx, 10)
	if err != nil {
		s.log.Error("failed to get unsent events", logger.Err(err))
		return
	}

	for _, event := range events {
		if err = s.producer.Produce(ctx, event.Topic, event.Payload); err != nil {
			s.log.Error("failed to publish event",
				slog.String("event_id", event.ID),
				slog.String("topic", event.Topic),
				logger.Err(err),
			)
			continue
		}

		if err = s.outbox.MarkSent(ctx, event.ID); err != nil {
			s.log.Error("failed to mark event as sent",
				slog.String("event_id", event.ID),
				logger.Err(err),
			)
		}
	}
}
