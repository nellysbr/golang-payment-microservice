package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"golang-payment-microservice/internal/model"
	"golang-payment-microservice/internal/queue"
	"golang-payment-microservice/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req *model.PaymentRequest) (*model.PaymentResponse, error)
	GetPayment(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	GetPaymentsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]*model.Payment, error)
	ProcessPaymentAsync(ctx context.Context, paymentID string) error
}

type paymentService struct {
	repo     repository.PaymentRepository
	producer queue.KafkaProducer
	logger   *logrus.Logger
}

func NewPaymentService(repo repository.PaymentRepository, producer queue.KafkaProducer, logger *logrus.Logger) PaymentService {
	return &paymentService{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (s *paymentService) CreatePayment(ctx context.Context, req *model.PaymentRequest) (*model.PaymentResponse, error) {
	// Validar dados do cartão
	card := &model.Card{
		Number:      req.CardNumber,
		Holder:      req.CardHolder,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
		CVV:         req.CVV,
	}

	if !card.IsValid() {
		return nil, fmt.Errorf("invalid card data")
	}

	// Verificar saldo da conta
	account, err := s.repo.GetAccountByCardNumber(ctx, req.CardNumber)
	if err != nil {
		s.logger.WithError(err).WithField("card_number", req.CardNumber).Error("Failed to get account")
		return nil, fmt.Errorf("account not found or invalid")
	}

	if !account.HasSufficientBalance(req.Amount) {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Criar pagamento
	payment := &model.Payment{
		ID:          uuid.New(),
		CardNumber:  req.CardNumber,
		CardHolder:  req.CardHolder,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
		CVV:         req.CVV,
		Amount:      req.Amount,
		Currency:    req.Currency,
		MerchantID:  req.MerchantID,
		Status:      model.PaymentStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Salvar no banco
	if err := s.repo.Create(ctx, payment); err != nil {
		s.logger.WithError(err).Error("Failed to create payment")
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Enviar para fila de processamento
	if err := s.producer.SendPaymentMessage(ctx, payment); err != nil {
		s.logger.WithError(err).WithField("payment_id", payment.ID).Error("Failed to send payment to queue")
		// Não retornar erro aqui, pois o pagamento foi criado
	}

	s.logger.WithField("payment_id", payment.ID).Info("Payment created successfully")

	return &model.PaymentResponse{
		ID:        payment.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		CreatedAt: payment.CreatedAt,
		Message:   "Payment created and queued for processing",
	}, nil
}

func (s *paymentService) GetPayment(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	payment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.WithError(err).WithField("payment_id", id).Error("Failed to get payment")
		return nil, err
	}

	return payment, nil
}

func (s *paymentService) GetPaymentsByMerchant(ctx context.Context, merchantID string, limit, offset int) ([]*model.Payment, error) {
	payments, err := s.repo.GetByMerchantID(ctx, merchantID, limit, offset)
	if err != nil {
		s.logger.WithError(err).WithField("merchant_id", merchantID).Error("Failed to get payments by merchant")
		return nil, err
	}

	return payments, nil
}

func (s *paymentService) ProcessPaymentAsync(ctx context.Context, paymentID string) error {
	id, err := uuid.Parse(paymentID)
	if err != nil {
		return fmt.Errorf("invalid payment ID: %w", err)
	}

	// Atualizar status para processando
	if err := s.repo.UpdateStatus(ctx, id, model.PaymentStatusProcessing, nil); err != nil {
		s.logger.WithError(err).WithField("payment_id", id).Error("Failed to update payment status to processing")
		return err
	}

	// Simular processamento (tempo aleatório entre 1-5 segundos)
	processingTime := time.Duration(rand.Intn(4)+1) * time.Second
	time.Sleep(processingTime)

	// Simular sucesso/falha (90% de sucesso)
	success := rand.Float32() < 0.9

	if success {
		// Processar pagamento com sucesso
		payment, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}

		// Debitar da conta
		account, err := s.repo.GetAccountByCardNumber(ctx, payment.CardNumber)
		if err != nil {
			errorMsg := "Failed to get account for debit"
			s.repo.UpdateStatus(ctx, id, model.PaymentStatusFailed, &errorMsg)
			return fmt.Errorf("failed to get account: %w", err)
		}

		newBalance := account.Balance - payment.Amount
		if err := s.repo.UpdateAccountBalance(ctx, payment.CardNumber, newBalance); err != nil {
			errorMsg := "Failed to update account balance"
			s.repo.UpdateStatus(ctx, id, model.PaymentStatusFailed, &errorMsg)
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// Atualizar status para completado
		if err := s.repo.UpdateStatus(ctx, id, model.PaymentStatusCompleted, nil); err != nil {
			return err
		}

		s.logger.WithField("payment_id", id).Info("Payment processed successfully")
	} else {
		// Simular falha no processamento
		errorMsg := "Payment processing failed due to external service error"
		if err := s.repo.UpdateStatus(ctx, id, model.PaymentStatusFailed, &errorMsg); err != nil {
			return err
		}

		s.logger.WithField("payment_id", id).Warn("Payment processing failed")
	}

	return nil
} 