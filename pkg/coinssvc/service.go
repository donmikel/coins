package coinssvc

import (
	"context"

	"github.com/donmikel/coins/pkg/coins"
	"github.com/donmikel/coins/pkg/payment"
	"github.com/go-kit/kit/log"
)

// Service provides payments functionality.
type Service interface {
	GetAvailableAccounts(ctx context.Context) (accounts []string, err error)
	GetAllPayments(ctx context.Context) (payments []payment.Payment, err error)
	SendPayment(ctx context.Context, input payment.PaymentInput) (err error)
}

type service struct {
	logger  log.Logger
	storage Storage
}

func newService(logger log.Logger, storage Storage) *service {
	return &service{
		logger:  logger,
		storage: storage,
	}
}

func (s *service) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	payments, err = s.storage.GetAllPayments(ctx)
	if err != nil {
		return nil, coins.ErrInternal("failed to get all payments: %s", err)
	}
	return
}

func (s *service) SendPayment(ctx context.Context, input payment.PaymentInput) (err error) {
	p := payment.Payment{
		FromAccount: input.FromAccount,
		ToAccount:   input.ToAccount,
		Direction:   input.Direction,
		Amount:      input.Amount,
	}
	err = s.storage.SendPayment(ctx, p)
	if err != nil {
		return coins.ErrInternal("failed to send payment: %s", err)
	}
	return
}

func (s *service) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	accounts, err = s.storage.GetAvailableAccounts(ctx)
	if err != nil {
		return nil, coins.ErrInternal("failed to get available accounts: %s", err)
	}
	return
}
