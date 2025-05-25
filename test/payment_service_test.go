package test

import (
	"context"
	"testing"
	"time"

	"golang-payment-microservice/internal/model"
	"golang-payment-microservice/internal/service"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus, errorMsg *string) error {
	args := m.Called(ctx, id, status, errorMsg)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByMerchantID(ctx context.Context, merchantID string, limit, offset int) ([]*model.Payment, error) {
	args := m.Called(ctx, merchantID, limit, offset)
	return args.Get(0).([]*model.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetAccountByCardNumber(ctx context.Context, cardNumber string) (*model.Account, error) {
	args := m.Called(ctx, cardNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Account), args.Error(1)
}

func (m *MockPaymentRepository) UpdateAccountBalance(ctx context.Context, cardNumber string, newBalance float64) error {
	args := m.Called(ctx, cardNumber, newBalance)
	return args.Error(0)
}

// Mock Kafka Producer
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) SendPaymentMessage(ctx context.Context, payment *model.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestPaymentService_CreatePayment_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockPaymentRepository)
	mockProducer := new(MockKafkaProducer)
	logger := logrus.New()
	
	paymentService := service.NewPaymentService(mockRepo, mockProducer, logger)

	// Mock data
	account := &model.Account{
		CardNumber: "1234567890123456",
		Balance:    1000.00,
		IsActive:   true,
	}

	req := &model.PaymentRequest{
		CardNumber:  "1234567890123456",
		CardHolder:  "John Doe",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		CVV:         "123",
		Amount:      100.00,
		Currency:    "BRL",
		MerchantID:  "merchant123",
	}

	// Setup expectations
	mockRepo.On("GetAccountByCardNumber", mock.Anything, req.CardNumber).Return(account, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.Payment")).Return(nil)
	mockProducer.On("SendPaymentMessage", mock.Anything, mock.AnythingOfType("*model.Payment")).Return(nil)

	// Execute
	response, err := paymentService.CreatePayment(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, model.PaymentStatusPending, response.Status)
	assert.Equal(t, req.Amount, response.Amount)
	assert.Equal(t, req.Currency, response.Currency)

	// Verify mocks
	mockRepo.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestPaymentService_CreatePayment_InsufficientBalance(t *testing.T) {
	// Setup
	mockRepo := new(MockPaymentRepository)
	mockProducer := new(MockKafkaProducer)
	logger := logrus.New()
	
	paymentService := service.NewPaymentService(mockRepo, mockProducer, logger)

	// Mock data
	account := &model.Account{
		CardNumber: "1234567890123456",
		Balance:    50.00, // Insufficient balance
		IsActive:   true,
	}

	req := &model.PaymentRequest{
		CardNumber:  "1234567890123456",
		CardHolder:  "John Doe",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		CVV:         "123",
		Amount:      100.00,
		Currency:    "BRL",
		MerchantID:  "merchant123",
	}

	// Setup expectations
	mockRepo.On("GetAccountByCardNumber", mock.Anything, req.CardNumber).Return(account, nil)

	// Execute
	response, err := paymentService.CreatePayment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "insufficient balance")

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestPaymentService_CreatePayment_InvalidCard(t *testing.T) {
	// Setup
	mockRepo := new(MockPaymentRepository)
	mockProducer := new(MockKafkaProducer)
	logger := logrus.New()
	
	paymentService := service.NewPaymentService(mockRepo, mockProducer, logger)

	req := &model.PaymentRequest{
		CardNumber:  "123", // Invalid card number
		CardHolder:  "John Doe",
		ExpiryMonth: 12,
		ExpiryYear:  2025,
		CVV:         "123",
		Amount:      100.00,
		Currency:    "BRL",
		MerchantID:  "merchant123",
	}

	// Execute
	response, err := paymentService.CreatePayment(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid card data")
}

func TestPaymentService_GetPayment_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockPaymentRepository)
	mockProducer := new(MockKafkaProducer)
	logger := logrus.New()
	
	paymentService := service.NewPaymentService(mockRepo, mockProducer, logger)

	// Mock data
	paymentID := uuid.New()
	payment := &model.Payment{
		ID:         paymentID,
		Amount:     100.00,
		Currency:   "BRL",
		Status:     model.PaymentStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Setup expectations
	mockRepo.On("GetByID", mock.Anything, paymentID).Return(payment, nil)

	// Execute
	result, err := paymentService.GetPayment(context.Background(), paymentID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, paymentID, result.ID)
	assert.Equal(t, payment.Amount, result.Amount)

	// Verify mocks
	mockRepo.AssertExpectations(t)
}

func TestCard_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		card     model.Card
		expected bool
	}{
		{
			name: "Valid card",
			card: model.Card{
				Number:      "1234567890123456",
				Holder:      "John Doe",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			expected: true,
		},
		{
			name: "Invalid card number length",
			card: model.Card{
				Number:      "123456789012345",
				Holder:      "John Doe",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
				CVV:         "123",
			},
			expected: false,
		},
		{
			name: "Expired card",
			card: model.Card{
				Number:      "1234567890123456",
				Holder:      "John Doe",
				ExpiryMonth: 1,
				ExpiryYear:  2020,
				CVV:         "123",
			},
			expected: false,
		},
		{
			name: "Invalid CVV",
			card: model.Card{
				Number:      "1234567890123456",
				Holder:      "John Doe",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
				CVV:         "12",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.card.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAccount_HasSufficientBalance(t *testing.T) {
	tests := []struct {
		name     string
		account  model.Account
		amount   float64
		expected bool
	}{
		{
			name: "Sufficient balance",
			account: model.Account{
				Balance:  1000.00,
				IsActive: true,
			},
			amount:   500.00,
			expected: true,
		},
		{
			name: "Insufficient balance",
			account: model.Account{
				Balance:  100.00,
				IsActive: true,
			},
			amount:   500.00,
			expected: false,
		},
		{
			name: "Inactive account",
			account: model.Account{
				Balance:  1000.00,
				IsActive: false,
			},
			amount:   500.00,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.account.HasSufficientBalance(tt.amount)
			assert.Equal(t, tt.expected, result)
		})
	}
} 