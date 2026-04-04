package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			BatchSize:    100,
			BatchTimeout: 50 * time.Millisecond,
			Async:        false, // sync для гарантии доставки аудит-логов
			Completion: func(messages []kafka.Message, err error) {
				if err != nil {
					slog.Error("Kafka producer error", "err", err)
				}
			},
		},
	}
}

// Publish отправляет событие в Kafka. Поддерживает контекст и отмену.
func (p *Producer) Publish(ctx context.Context, topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal kafka payload: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: data,
		Time:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("kafka write message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}