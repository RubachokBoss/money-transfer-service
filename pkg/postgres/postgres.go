package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func Connect(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	var balance float64
	err := r.db.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1", id).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("error getting balance: %w", err)
	}
	return balance, nil
}

func (r *PostgresRepository) TransferMoney(ctx context.Context, from, to uuid.UUID, amount float64, currency string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentBalance float64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", from).Scan(&currentBalance)
	if err != nil {
		return err
	}
	if currentBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
	if err != nil {
		return err
	}

	// Добавьте запись в таблицу transfers
	_, err = tx.ExecContext(ctx, `
        INSERT INTO transfers (from_account_id, to_account_id, amount, currency)
        VALUES ($1, $2, $3, $4)
    `, from, to, amount, currency)
	if err != nil {
		return err
	}

	return tx.Commit()
}
