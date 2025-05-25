package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-payment-microservice/config"
	"golang-payment-microservice/internal/handler"
	"golang-payment-microservice/internal/queue"
	"golang-payment-microservice/internal/repository"
	"golang-payment-microservice/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configurar logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Carregar configurações
	cfg := config.Load()
	logger.Info("Configuration loaded successfully")

	// Conectar ao banco de dados
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer dbPool.Close()

	// Verificar conexão com o banco
	if err := dbPool.Ping(context.Background()); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}
	logger.Info("Database connection established")

	// Inicializar repositório
	paymentRepo := repository.NewPaymentRepository(dbPool)

	// Inicializar produtor Kafka
	kafkaProducer := queue.NewKafkaProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, logger)
	defer kafkaProducer.Close()

	// Inicializar serviço
	paymentService := service.NewPaymentService(paymentRepo, kafkaProducer, logger)

	// Inicializar consumidor Kafka
	kafkaConsumer := queue.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, "payment-processor", paymentService, logger)

	// Inicializar handler HTTP
	httpHandler := handler.NewHTTPHandler(paymentService, logger)
	router := httpHandler.SetupRoutes()

	// Servidor HTTP
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.HTTPPort),
		Handler: router,
	}

	// Servidor de métricas
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Metrics.Port),
		Handler: promhttp.Handler(),
	}

	// Canal para capturar sinais do sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidores em goroutines
	go func() {
		logger.WithField("port", cfg.Server.HTTPPort).Info("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	go func() {
		logger.WithField("port", cfg.Metrics.Port).Info("Starting metrics server")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start metrics server")
		}
	}()

	// Iniciar consumidor Kafka
	go func() {
		logger.Info("Starting Kafka consumer")
		if err := kafkaConsumer.Start(context.Background()); err != nil {
			logger.WithError(err).Error("Kafka consumer stopped with error")
		}
	}()

	logger.Info("Payment microservice started successfully")

	// Aguardar sinal de parada
	<-quit
	logger.Info("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parar servidores
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server forced to shutdown")
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Metrics server forced to shutdown")
	}

	// Fechar consumidor Kafka
	if err := kafkaConsumer.Close(); err != nil {
		logger.WithError(err).Error("Failed to close Kafka consumer")
	}

	logger.Info("Payment microservice stopped")
}