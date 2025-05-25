package repository

import (
	"context"
	"fmt"
	"time"

	"golang-payment-microservice/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus, errorMsg *string) error
	GetByMerchantID(ctx context.Context, merchantID string, limit, offset int) ([]*model.Payment, error)
	GetAccountByCardNumber(ctx context.Context, cardNumber string) (*model.Account, error)
	UpdateAccountBalance(ctx context.Context, cardNumber string, newBalance float64) error
}

type paymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *model.Payment) error {
	query := `
		INSERT INTO payments (
			id, card_number, card_holder, expiry_month, expiry_year, 
			cvv, amount, currency, merchant_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	
	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.CardNumber,
		payment.CardHolder,
		payment.ExpiryMonth,
		payment.ExpiryYear,
		payment.CVV,
		payment.Amount,
		payment.Currency,
		payment.MerchantID,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	
	return err
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	query := `
		SELECT id, card_number, card_holder, expiry_month, expiry_year,
			   cvv, amount, currency, merchant_id, status, created_at, 
			   updated_at, processed_at, error_msg
		FROM payments 
		WHERE id = $1
	`
	
	payment := &model.Payment{}
	row := r.db.QueryRow(ctx, query, id)
	
	err := row.Scan(
		&payment.ID,
		&payment.CardNumber,
		&payment.CardHolder,
		&payment.ExpiryMonth,
		&payment.ExpiryYear,
		&payment.CVV,
		&payment.Amount,
		&payment.Currency,
		&payment.MerchantID,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.ProcessedAt,
		&payment.ErrorMsg,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, err
	}
	
	return payment, nil
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.PaymentStatus, errorMsg *string) error {
	query := `
		UPDATE payments 
		SET status = $2, updated_at = $3, processed_at = $4, error_msg = $5
		WHERE id = $1
	`
	
	now := time.Now()
	_, err := r.db.Exec(ctx, query, id, status, now, now, errorMsg)
	return err
}

func (r *paymentRepository) GetByMerchantID(ctx context.Context, merchantID string, limit, offset int) ([]*model.Payment, error) {
	query := `
		SELECT id, card_number, card_holder, expiry_month, expiry_year,
			   cvv, amount, currency, merchant_id, status, created_at, 
			   updated_at, processed_at, error_msg
		FROM payments 
		WHERE merchant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := r.db.Query(ctx, query, merchantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var payments []*model.Payment
	for rows.Next() {
		payment := &model.Payment{}
		err := rows.Scan(
			&payment.ID,
			&payment.CardNumber,
			&payment.CardHolder,
			&payment.ExpiryMonth,
			&payment.ExpiryYear,
			&payment.CVV,
			&payment.Amount,
			&payment.Currency,
			&payment.MerchantID,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&payment.ProcessedAt,
			&payment.ErrorMsg,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	
	return payments, rows.Err()
}

func (r *paymentRepository) GetAccountByCardNumber(ctx context.Context, cardNumber string) (*model.Account, error) {
	query := `
		SELECT card_number, balance, is_active, created_at, updated_at
		FROM accounts 
		WHERE card_number = $1
	`
	
	account := &model.Account{}
	row := r.db.QueryRow(ctx, query, cardNumber)
	
	err := row.Scan(
		&account.CardNumber,
		&account.Balance,
		&account.IsActive,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}
	
	return account, nil
}

func (r *paymentRepository) UpdateAccountBalance(ctx context.Context, cardNumber string, newBalance float64) error {
	query := `
		UPDATE accounts 
		SET balance = $2, updated_at = $3
		WHERE card_number = $1
	`
	
	_, err := r.db.Exec(ctx, query, cardNumber, newBalance, time.Now())
	return err
} 