package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// NewKafkaWriter создает Kafka Writer с переданными параметрами
func NewKafkaWriter(
	brokers []string,
	topic string,
	requiredAcks string,
	batchTimeout int64,
	batchSize int,
	allowAutoTopicCreation bool,
	logger *logrus.Logger,
) (*kafka.Writer, error) {
	// Преобразуем строковое значение для RequiredAcks
	var acks kafka.RequiredAcks
	switch requiredAcks {
	case "one":
		acks = kafka.RequireOne
	case "all":
		acks = kafka.RequireAll
	default:
		acks = kafka.RequireAll // По умолчанию все брокеры подтверждают запись
	}

	// Настройки Kafka Writer
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           acks,
		BatchTimeout:           time.Duration(batchTimeout) * time.Millisecond,
		BatchSize:              batchSize,
		AllowAutoTopicCreation: allowAutoTopicCreation,
	}

	// Тестовое сообщение для проверки соединения
	testMessage := kafka.Message{Value: []byte("test")}
	if err := writer.WriteMessages(context.Background(), testMessage); err != nil {
		return nil, fmt.Errorf("failed to write test message to Kafka: %w", err)
	}

	logger.Debug("Kafka writer initialized successfully")
	return writer, nil
}
