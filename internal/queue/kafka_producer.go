package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"golang-payment-microservice/internal/model"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type PaymentMessage struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Timestamp int64   `json:"timestamp"`
}

type KafkaProducer interface {
	SendPaymentMessage(ctx context.Context, payment *model.Payment) error
	Close() error
}

type kafkaProducer struct {
	writer *kafka.Writer
	logger *logrus.Logger
}

func NewKafkaProducer(brokers []string, topic string, logger *logrus.Logger) KafkaProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &kafkaProducer{
		writer: writer,
		logger: logger,
	}
}

func (p *kafkaProducer) SendPaymentMessage(ctx context.Context, payment *model.Payment) error {
	message := PaymentMessage{
		PaymentID: payment.ID.String(),
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		Timestamp: payment.CreatedAt.Unix(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		p.logger.WithError(err).Error("Failed to marshal payment message")
		return fmt.Errorf("failed to marshal payment message: %w", err)
	}

	kafkaMessage := kafka.Message{
		Key:   []byte(payment.ID.String()),
		Value: messageBytes,
	}

	err = p.writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		p.logger.WithError(err).WithField("payment_id", payment.ID).Error("Failed to send payment message to Kafka")
		return fmt.Errorf("failed to send payment message: %w", err)
	}

	p.logger.WithField("payment_id", payment.ID).Info("Payment message sent to Kafka successfully")
	return nil
}

func (p *kafkaProducer) Close() error {
	return p.writer.Close()
} 