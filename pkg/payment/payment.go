package payment

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// Direction is possible payment direction
type Direction uint16

const (
	Incomming Direction = iota
	Outgoing
)

// Payment is a single payment transaction from user to user.
type Payment struct {
	ID          uint64          `json:"id" db:"id"`
	FromAccount string          `json:"from_account" db:"from_account"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	ToAccount   string          `json:"to_account" db:"to_account"`
	Direction   Direction       `json:"direction" db:"direction"`
	Dt          *time.Time      `json:"dt" db:"dt"`
}

// PaymentInput is an input structure used to create new payment aka send payment.
type PaymentInput struct {
	FromAccount string          `json:"from_account"`
	Amount      decimal.Decimal `json:"amount"`
	ToAccount   string          `json:"to_account"`
	Direction   Direction       `json:"direction"`
}

// Validate validates the given Payment structure
func (p Payment) Validate() error {
	if p.FromAccount == "" {
		return errors.New("empty FromAccount")
	}
	if p.ToAccount == "" {
		return errors.New("empty ToAccount")
	}
	if p.Amount.LessThanOrEqual(decimal.NewFromInt(0)) {
		return errors.New("invalid Amount")
	}

	return nil
}
