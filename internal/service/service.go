package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"money-transfer-service/internal/cache"
	"money-transfer-service/internal/models"
	"money-transfer-service/internal/repository"

	"github.com/google/uuid"
)

type Service struct {
	repo  *repository.Repository
	cache *cache.RedisClient
}

func (s *Service) GetTransfersHistory(ctx context.Context, accountID uuid.UUID) ([]models.Transfer, error) {
	return s.repo.GetTransfersByAccount(ctx, accountID)
}
func NewService(repo *repository.Repository, cache *cache.RedisClient) *Service {
	return &Service{repo: repo, cache: cache}
}

func (s *Service) GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	return s.repo.GetBalance(ctx, accountID)
}
func (s *Service) DepositMoney(ctx context.Context, accountID uuid.UUID, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.DepositMoney(ctx, accountID, amount)
}
func (s *Service) TransferMoney(ctx context.Context, req models.TransferRequest) error {
	fromID, err := uuid.Parse(req.From)
	if err != nil {
		return fmt.Errorf("invalid from account ID")
	}
	toID, err := uuid.Parse(req.To)
	if err != nil {
		return fmt.Errorf("invalid to account ID")
	}

	// Для конвертации используем исходную сумму и валюту
	amountToTransfer := req.Amount
	if req.Currency != "RUB" {
		rate, err := s.getExchangeRate(ctx, req.Currency)
		if err != nil {
			return fmt.Errorf("failed to get exchange rate: %w", err)
		}
		amountToTransfer = req.Amount * rate
	}

	return s.repo.TransferMoney(ctx, fromID, toID, amountToTransfer, req.Currency)
}

func (s *Service) TransferMoneyByEmail(ctx context.Context, fromUserID uuid.UUID, toEmail string, amount float64, currency string) error {
	fromAccount, err := s.repo.GetAccountByUserID(ctx, fromUserID)
	if err != nil {
		return err
	}
	if fromAccount == nil {
		return fmt.Errorf("sender account not found")
	}

	toAccount, err := s.repo.GetAccountByEmail(ctx, toEmail)
	if err != nil {
		return err
	}
	if toAccount == nil {
		return fmt.Errorf("recipient account not found for email: %s", toEmail)
	}

	// Для конвертации используем исходную сумму и валюту
	amountToTransfer := amount
	if currency != "RUB" {
		rate, err := s.getExchangeRate(ctx, currency)
		if err != nil {
			return fmt.Errorf("failed to get exchange rate: %w", err)
		}
		amountToTransfer = amount * rate
	}

	return s.repo.TransferMoney(ctx, fromAccount.ID, toAccount.ID, amountToTransfer, currency)
}

func (s *Service) getExchangeRate(ctx context.Context, currency string) (float64, error) {
	cachedRate, err := s.cache.Get(ctx, currency)
	if err == nil {
		rate, err := strconv.ParseFloat(cachedRate, 64)
		if err == nil {
			return rate, nil
		}
	}

	// Актуальные курсы (1 USD/EUR к RUB)
	rates := map[string]float64{
		"USD": 90.0,  // 1 USD = 90 RUB
		"EUR": 100.0, // 1 EUR = 100 RUB
	}

	rate, ok := rates[currency]
	if !ok {
		return 0, fmt.Errorf("currency not supported: %s", currency)
	}

	// Кэшируем на 5 минут
	err = s.cache.Set(ctx, currency, fmt.Sprintf("%f", rate), 5*time.Minute)
	if err != nil {
		log.Printf("Cache set error: %v", err)
	}

	return rate, nil
}
func (s *Service) GetAccountByUserID(ctx context.Context, userID uuid.UUID) (*models.Account, error) {
	return s.repo.GetAccountByUserID(ctx, userID)
}
