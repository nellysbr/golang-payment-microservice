package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

// PaymentProcessor interface para evitar dependência circular
type PaymentProcessor interface {
	ProcessPaymentAsync(ctx context.Context, paymentID string) error
}

type KafkaConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

type kafkaConsumer struct {
	reader           *kafka.Reader
	paymentProcessor PaymentProcessor
	logger           *logrus.Logger
}

func NewKafkaConsumer(brokers []string, topic, groupID string, paymentProcessor PaymentProcessor, logger *logrus.Logger) KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &kafkaConsumer{
		reader:           reader,
		paymentProcessor: paymentProcessor,
		logger:           logger,
	}
}

func (c *kafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting Kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Kafka consumer stopped")
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.WithError(err).Error("Failed to read message from Kafka")
				continue
			}

			c.processMessage(ctx, message)
		}
	}
}

func (c *kafkaConsumer) processMessage(_ context.Context, message kafka.Message) {
	var paymentMsg PaymentMessage
	if err := json.Unmarshal(message.Value, &paymentMsg); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal payment message")
		return
	}

	c.logger.WithField("payment_id", paymentMsg.PaymentID).Info("Processing payment message")

	// Simular processamento assíncrono
	go func() {
		processingCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := c.paymentProcessor.ProcessPaymentAsync(processingCtx, paymentMsg.PaymentID); err != nil {
			c.logger.WithError(err).WithField("payment_id", paymentMsg.PaymentID).Error("Failed to process payment")
		} else {
			c.logger.WithField("payment_id", paymentMsg.PaymentID).Info("Payment processed successfully")
		}
	}()
}

func (c *kafkaConsumer) Close() error {
	return c.reader.Close()
} 