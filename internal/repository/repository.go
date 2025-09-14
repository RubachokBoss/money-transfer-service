package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"money-transfer-service/internal/models"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func (r *Repository) CreateTransfer(ctx context.Context, from, to uuid.UUID, amount float64, currency string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO transfers (from_account_id, to_account_id, amount, currency)
        VALUES ($1, $2, $3, $4)
    `, from, to, amount, currency)
	return err
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
func (r *Repository) DepositMoney(ctx context.Context, accountID uuid.UUID, amount float64) error {
	log.Printf("Attempting to deposit %.2f to account %s", amount, accountID.String())

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверим существование счета
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)", accountID).Scan(&exists)
	if err != nil {
		log.Printf("Error checking account existence: %v", err)
		return err
	}

	if !exists {
		return fmt.Errorf("account not found")
	}

	// Выполним пополнение
	result, err := tx.ExecContext(ctx, `
        UPDATE accounts 
        SET balance = balance + $1 
        WHERE id = $2
    `, amount, accountID)

	if err != nil {
		log.Printf("Deposit error: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Rows affected: %d", rowsAffected)

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	// Зафиксируем транзакцию
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Printf("Deposit successful")
	return nil
}

// Добавляем методы интерфейса
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, full_name, created_at 
        FROM users WHERE email = $1
    `, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
        SELECT id, email, password_hash, full_name, created_at 
        FROM users WHERE id = $1
    `, id).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO users (email, password_hash, full_name) 
        VALUES ($1, $2, $3)
        RETURNING id, email, password_hash, full_name, created_at
    `, email, passwordHash, fullName).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateAccount(ctx context.Context, userID uuid.UUID) (*models.Account, error) {
	var account models.Account
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO accounts (user_id, balance) 
        VALUES ($1, $2)
        RETURNING id, user_id, balance
    `, userID, 0.00).Scan(&account.ID, &account.UserID, &account.Balance)

	if err != nil {
		return nil, err
	}
	return &account, nil
}
func (r *Repository) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*models.Account, error) {
	var account models.Account
	err := r.db.QueryRowContext(ctx, `
        SELECT id, user_id, balance 
        FROM accounts 
        WHERE user_id = $1
    `, userID).Scan(&account.ID, &account.UserID, &account.Balance)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	var account models.Account
	err := r.db.QueryRowContext(ctx, `
        SELECT a.id, a.user_id, a.balance 
        FROM accounts a
        JOIN users u ON a.user_id = u.id
        WHERE u.email = $1
    `, email).Scan(&account.ID, &account.UserID, &account.Balance)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) GetTransfersByAccount(ctx context.Context, accountID uuid.UUID) ([]models.Transfer, error) {
	query := `
        SELECT t.id, t.from_account_id, t.to_account_id, t.amount, t.currency, t.created_at,
               u1.email as from_email, u2.email as to_email
        FROM transfers t
        LEFT JOIN accounts a1 ON t.from_account_id = a1.id
        LEFT JOIN users u1 ON a1.user_id = u1.id
        LEFT JOIN accounts a2 ON t.to_account_id = a2.id
        LEFT JOIN users u2 ON a2.user_id = u2.id
        WHERE t.from_account_id = $1 OR t.to_account_id = $1
        ORDER BY t.created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transfers []models.Transfer
	for rows.Next() {
		var t models.Transfer
		err := rows.Scan(
			&t.ID,
			&t.From,
			&t.To,
			&t.Amount,
			&t.Currency,
			&t.CreatedAt,
			&t.FromEmail,
			&t.ToEmail,
		)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, t)
	}
	return transfers, nil
}

// Остальные методы...

func (r *Repository) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	// Реализация будет в postgres.go
	return 0, nil
}

func (r *Repository) TransferMoney(ctx context.Context, from, to uuid.UUID, amount float64, currency string) error {
	log.Printf("Transfer attempt: from=%s, to=%s, amount=%.2f, currency=%s", from, to, amount, currency)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Begin transaction error: %v", err)
		return err
	}
	defer tx.Rollback()

	// Проверяем баланс отправителя в RUB
	var currentBalance float64
	err = tx.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = $1 FOR UPDATE", from).Scan(&currentBalance)
	if err != nil {
		log.Printf("Balance check error: %v", err)
		return err
	}

	log.Printf("Current balance: %.2f, Transfer amount: %.2f", currentBalance, amount)

	if currentBalance < amount {
		log.Printf("Insufficient funds: have %.2f, need %.2f", currentBalance, amount)
		return fmt.Errorf("insufficient funds")
	}

	// Списание средств
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, from)
	if err != nil {
		log.Printf("Debit error: %v", err)
		return err
	}

	// Зачисление средств
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, to)
	if err != nil {
		log.Printf("Credit error: %v", err)
		return err
	}

	// Запись о переводе
	_, err = tx.ExecContext(ctx, `
        INSERT INTO transfers (from_account_id, to_account_id, amount, currency)
        VALUES ($1, $2, $3, $4)
    `, from, to, amount, currency)
	if err != nil {
		log.Printf("Transfer record error: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Commit error: %v", err)
		return err
	}

	log.Printf("Transfer completed successfully")
	return nil
}
