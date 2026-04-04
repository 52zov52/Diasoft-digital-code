package kafka

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	kafka "github.com/segmentio/kafka-go"
)

type ConsumerHandler func(ctx context.Context, msg []byte) error

type Consumer struct {
	reader  *kafka.Reader
	handler ConsumerHandler
}

func NewConsumer(brokers []string, topic, groupID string, handler ConsumerHandler) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		handler: handler,
	}
}

// Start запускает цикл чтения. Грейсфульно останавливается по сигналу.
func (c *Consumer) Start(ctx context.Context) {
	defer c.reader.Close()
	
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					slog.Info("Kafka consumer stopped gracefully")
					return
				}
				slog.Error("Kafka read error", "err", err)
				continue
			}

			if err := c.handler(ctx, m.Value); err != nil {
				slog.Error("Kafka handler error", "topic", m.Topic, "err", err)
				// В проде: dead letter queue или повторный пуш
			}
		}
	}()

	<-sig
	slog.Info("Received shutdown signal, stopping consumer...")
}