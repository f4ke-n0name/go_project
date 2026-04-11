package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	log    *slog.Logger
	client *kgo.Client
}

func NewProducer(log *slog.Logger, brokers []string) (*Producer, error) {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.AllowAutoTopicCreation())

	if err != nil {
		return nil, fmt.Errorf("cannot create kafka producer: %w", err)
	}

	if err = client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("cannot ping kafka producer: %w", err)
	}
	log.Info("kafka producer created", slog.String("broker", strings.Join(brokers, ",")))

	return &Producer{log: log, client: client}, nil
}

func (p *Producer) Produce(ctx context.Context, topic string, payload []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Value: payload,
	}

	if err := p.client.ProduceSync(ctx, record).FirstErr(); err != nil {
		p.log.Error("failed to publish message",
			slog.String("topic", topic),
			logger.Err(err),
		)
		return fmt.Errorf("publish to %s: %w", topic, err)
	}
	p.log.Debug("message published", slog.String("topic", topic))
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}
