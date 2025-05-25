package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Contador de pagamentos criados
	PaymentsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_created_total",
			Help: "Total number of payments created",
		},
		[]string{"merchant_id", "currency"},
	)

	// Contador de pagamentos processados por status
	PaymentsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payments_processed_total",
			Help: "Total number of payments processed by status",
		},
		[]string{"status", "merchant_id"},
	)

	// Histograma do tempo de processamento
	PaymentProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_processing_duration_seconds",
			Help:    "Time taken to process payments",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	// Gauge do valor total de pagamentos
	PaymentAmountTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_amount_total",
			Help: "Total amount of payments",
		},
		[]string{"currency"},
	)

	// Contador de erros HTTP
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Histograma da duração das requisições HTTP
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Gauge de conexões ativas do banco
	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "database_connections_active",
			Help: "Number of active database connections",
		},
	)

	// Contador de mensagens Kafka
	KafkaMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_total",
			Help: "Total number of Kafka messages",
		},
		[]string{"topic", "operation", "status"},
	)
)

// RecordPaymentCreated registra a criação de um pagamento
func RecordPaymentCreated(merchantID, currency string) {
	PaymentsCreatedTotal.WithLabelValues(merchantID, currency).Inc()
}

// RecordPaymentProcessed registra o processamento de um pagamento
func RecordPaymentProcessed(status, merchantID string) {
	PaymentsProcessedTotal.WithLabelValues(status, merchantID).Inc()
}

// RecordPaymentAmount registra o valor de um pagamento
func RecordPaymentAmount(currency string, amount float64) {
	PaymentAmountTotal.WithLabelValues(currency).Add(amount)
}

// RecordHTTPRequest registra uma requisição HTTP
func RecordHTTPRequest(method, endpoint, statusCode string) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// RecordKafkaMessage registra uma mensagem Kafka
func RecordKafkaMessage(topic, operation, status string) {
	KafkaMessagesTotal.WithLabelValues(topic, operation, status).Inc()
} 