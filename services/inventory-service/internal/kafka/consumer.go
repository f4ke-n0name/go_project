package kafka

import (
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
)

const workerPoolSize = 5

type OrderHandler interface {
	HandleOrderCreated(ctx context.Context, payload []byte) error
	HandleOrderStatusChanged(ctx context.Context, payload []byte) error
}

type Consumer struct {
	log     *slog.Logger
	client  *kgo.Client
	handler OrderHandler
}

func NewConsumer(log *slog.Logger, brokers []string, groupID string, handler OrderHandler) (*Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics("order.created", "order.status_changed"),
	)
	if err != nil {
		return nil, err
	}

	log.Info("kafka consumer created",
		slog.Any("brokers", brokers),
		slog.String("group_id", groupID),
	)

	return &Consumer{log: log, client: client, handler: handler}, nil
}

func (c *Consumer) Close() {
	c.client.Close()
}

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("starting kafka consumer", slog.Int("workers", workerPoolSize))

	jobs := make(chan *kgo.Record, 100)

	for i := 0; i < workerPoolSize; i++ {
		go func(workerID int) {
			c.log.Debug("worker started", slog.Int("worker_id", workerID))
			for record := range jobs {
				c.processRecord(ctx, record)
			}
		}(i)
	}

	go func() {
		defer close(jobs)
		for {
			select {
			case <-ctx.Done():
				c.log.Info("kafka consumer stopped")
				return
			default:
				fetches := c.client.PollFetches(ctx)
				if fetches.IsClientClosed() {
					return
				}

				fetches.EachError(func(t string, p int32, err error) {
					c.log.Error("kafka fetch error",
						slog.String("topic", t),
						slog.Int("partition", int(p)),
						logger.Err(err),
					)
				})

				fetches.EachRecord(func(record *kgo.Record) {
					jobs <- record
				})
			}
		}
	}()
}

func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) {
	log := c.log.With(
		slog.String("topic", record.Topic),
		slog.Int("partition", int(record.Partition)),
	)

	var err error
	switch record.Topic {
	case "order.created":
		err = c.handler.HandleOrderCreated(ctx, record.Value)
	case "order.status_changed":
		err = c.handler.HandleOrderStatusChanged(ctx, record.Value)
	default:
		log.Warn("unknown topic")
	}

	if err != nil {
		log.Error("failed to process record", logger.Err(err))
		return
	}

	c.client.MarkCommitRecords(record)
}
