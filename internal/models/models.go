package models

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID      uuid.UUID `json:"id" db:"id"`
	UserID  uuid.UUID `json:"user_id" db:"user_id"` // Добавьте это поле
	Balance float64   `json:"balance" db:"balance"`
}

type TransferRequest struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"` // Например: "USD", "EUR"
}

type Transfer struct {
	ID        uuid.UUID `json:"id" db:"id"`
	From      uuid.UUID `json:"from_account_id" db:"from_account_id"`
	To        uuid.UUID `json:"to_account_id" db:"to_account_id"`
	FromEmail string    `json:"from_email" db:"from_email"`
	ToEmail   string    `json:"to_email" db:"to_email"`
	Amount    float64   `json:"amount" db:"amount"`
	Currency  string    `json:"currency" db:"currency"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
