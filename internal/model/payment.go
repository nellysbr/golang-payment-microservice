package model

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus representa o status de um pagamento
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment representa uma transação de pagamento
type Payment struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	CardNumber  string        `json:"card_number" db:"card_number"`
	CardHolder  string        `json:"card_holder" db:"card_holder"`
	ExpiryMonth int           `json:"expiry_month" db:"expiry_month"`
	ExpiryYear  int           `json:"expiry_year" db:"expiry_year"`
	CVV         string        `json:"cvv" db:"cvv"`
	Amount      float64       `json:"amount" db:"amount"`
	Currency    string        `json:"currency" db:"currency"`
	MerchantID  string        `json:"merchant_id" db:"merchant_id"`
	Status      PaymentStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty" db:"processed_at"`
	ErrorMsg    *string       `json:"error_msg,omitempty" db:"error_msg"`
}

// PaymentRequest representa uma solicitação de pagamento
type PaymentRequest struct {
	CardNumber  string  `json:"card_number" validate:"required,len=16"`
	CardHolder  string  `json:"card_holder" validate:"required,min=3,max=100"`
	ExpiryMonth int     `json:"expiry_month" validate:"required,min=1,max=12"`
	ExpiryYear  int     `json:"expiry_year" validate:"required,min=2024"`
	CVV         string  `json:"cvv" validate:"required,len=3"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	MerchantID  string  `json:"merchant_id" validate:"required"`
}

// PaymentResponse representa a resposta de uma solicitação de pagamento
type PaymentResponse struct {
	ID        uuid.UUID     `json:"id"`
	Status    PaymentStatus `json:"status"`
	Amount    float64       `json:"amount"`
	Currency  string        `json:"currency"`
	CreatedAt time.Time     `json:"created_at"`
	Message   string        `json:"message,omitempty"`
}

// Card representa informações de um cartão
type Card struct {
	Number      string `json:"number"`
	Holder      string `json:"holder"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

// IsValid verifica se o cartão é válido (validação básica)
func (c *Card) IsValid() bool {
	// Validação básica do número do cartão (Luhn algorithm seria ideal)
	if len(c.Number) != 16 {
		return false
	}
	
	// Validação da data de expiração
	currentYear := time.Now().Year()
	currentMonth := int(time.Now().Month())
	
	if c.ExpiryYear < currentYear {
		return false
	}
	
	if c.ExpiryYear == currentYear && c.ExpiryMonth < currentMonth {
		return false
	}
	
	// Validação do CVV
	if len(c.CVV) != 3 {
		return false
	}
	
	return true
}

// Account representa uma conta simulada para validação de saldo
type Account struct {
	CardNumber string  `json:"card_number" db:"card_number"`
	Balance    float64 `json:"balance" db:"balance"`
	IsActive   bool    `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// HasSufficientBalance verifica se a conta tem saldo suficiente
func (a *Account) HasSufficientBalance(amount float64) bool {
	return a.IsActive && a.Balance >= amount
} 