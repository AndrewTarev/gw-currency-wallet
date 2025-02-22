package kafka

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
)

const (
	maxRetries    = 5                  // Максимальное число ретраев
	retryInterval = 2 * time.Second    // Интервал между попытками
	dlqTopic      = "transactions-dlq" // Dead Letter Queue (DLQ)
)

type KafkaConsumer struct {
	reader   *kafka.Reader
	producer *kafka.Writer
	metrics  *ConsumerMetrics
}

// ConsumerMetrics - структура для Prometheus метрик
type ConsumerMetrics struct {
	processedMessages prometheus.Counter
	failedMessages    prometheus.Counter
}

func NewConsumerMetrics() *ConsumerMetrics {
	m := &ConsumerMetrics{
		processedMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "consumer_processed_messages",
			Help: "Total number of processed messages",
		}),
		failedMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "consumer_failed_messages",
			Help: "Total number of failed messages",
		}),
	}
	prometheus.MustRegister(m.processedMessages, m.failedMessages)
	return m
}

// NewKafkaConsumer создаёт новый Kafka Consumer с поддержкой DLQ и метрик
func NewKafkaConsumer(brokers []string, topic, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     groupID,
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
			MaxWait:     500 * time.Millisecond,
			StartOffset: kafka.FirstOffset,
		}),
		producer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        dlqTopic,
			BatchSize:    1, // Отправляем сразу после ошибки
			BatchTimeout: 10 * time.Millisecond,
		},
		metrics: NewConsumerMetrics(),
	}
}

// processMessage обработка сообщения
func (kc *KafkaConsumer) processMessage(msg kafka.Message) error {
	// ❌ Симуляция ошибки для теста
	if string(msg.Value) == "error" {
		return fmt.Errorf("simulated error")
	}

	// ✅ Сообщение обработано успешно
	log.Printf("✅ Processed message: %s", string(msg.Value))
	kc.metrics.processedMessages.Inc()
	return nil
}

// retryProcessing выполняет повторную обработку с ретраями
func (kc *KafkaConsumer) retryProcessing(ctx context.Context, msg kafka.Message) {
	for i := 0; i < maxRetries; i++ {
		err := kc.processMessage(msg)
		if err == nil {
			return // ✅ Если успешно, выходим
		}
		log.Printf("🔄 Retry %d/%d: %v\n", i+1, maxRetries, err)
		time.Sleep(retryInterval)
	}

	// ❌ После maxRetries отправляем в DLQ
	err := kc.producer.WriteMessages(ctx, kafka.Message{
		Key:   msg.Key,
		Value: msg.Value,
	})
	if err != nil {
		log.Printf("❌ Failed to send to DLQ: %v\n", err)
	} else {
		log.Printf("☠️ Message sent to DLQ: %s\n", string(msg.Value))
	}
	kc.metrics.failedMessages.Inc()
}

// Start запускает consumer
func (kc *KafkaConsumer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			msg, err := kc.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("❌ Error reading message: %v\n", err)
				continue
			}

			// Обрабатываем с ретраями
			kc.retryProcessing(ctx, msg)
		}
	}()

	// Обработка SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("🛑 Shutting down consumer...")
	kc.reader.Close()
	kc.producer.Close()
}
