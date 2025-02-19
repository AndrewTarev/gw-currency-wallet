package logging

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/IBM/sarama"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func SetupLogger(level, format, outputFile, kafkaTopic string, kafkaBrokers []string) (*logrus.Logger, error) {
	logger := logrus.New()

	// Устанавливаем уровень логирования
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.Warnf("Не удалось установить уровень логирования '%s', используется INFO", level)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Настраиваем формат логирования
	switch format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	default:
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Настраиваем вывод логов (консоль + файл с ротацией)
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	if outputFile != "" {
		logWriter, err := rotatelogs.New(
			outputFile+".%Y-%m-%d",
			rotatelogs.WithMaxAge(7*24*time.Hour),     // Храним логи 7 дней
			rotatelogs.WithRotationTime(24*time.Hour), // Ротация раз в сутки
		)
		if err != nil {
			return nil, errors.Wrap(err, "ошибка настройки ротации логов")
		}
		writers = append(writers, logWriter)
	}

	// Устанавливаем многопоточный вывод логов
	logger.SetOutput(io.MultiWriter(writers...))

	// Добавляем хук для stack trace
	logger.AddHook(&StackTraceHook{})

	// Настроим Kafka логирование
	if len(kafkaBrokers) > 0 && kafkaTopic != "" {
		producer, err := newKafkaProducer(kafkaBrokers)
		if err != nil {
			return nil, errors.Wrap(err, "ошибка настройки Kafka логирования")
		}
		logger.AddHook(&kafkaHook{
			producer: producer,
			topic:    kafkaTopic,
		})
	}

	return logger, nil
}

// StackTraceHook - добавляет stack trace к ошибкам
type StackTraceHook struct{}

func (h *StackTraceHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func (h *StackTraceHook) Fire(entry *logrus.Entry) error {
	// Если в логах уже есть stack trace — ничего не делаем
	if _, exists := entry.Data["stack_trace"]; !exists {
		entry.Data["stack_trace"] = GetStackTrace()
	}
	return nil
}

func GetStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// newKafkaProducer - инициализация Kafka producer
func newKafkaProducer(brokers []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	return sarama.NewSyncProducer(brokers, config)
}

// kafkaHook - хук для отправки логов в Kafka
type kafkaHook struct {
	producer sarama.SyncProducer
	topic    string
}

// Fire - отправка логов в Kafka
func (h *kafkaHook) Fire(entry *logrus.Entry) error {
	message, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}
	kafkaMsg := &sarama.ProducerMessage{
		Topic: h.topic,
		Value: sarama.StringEncoder(message),
	}
	_, _, err = h.producer.SendMessage(kafkaMsg)
	return err
}

func (h *kafkaHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
