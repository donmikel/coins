package account

import (
	"errors"

	"github.com/shopspring/decimal"
)

// Account is single account
type Account struct {
	ID       string          `json:"id" db:"id"`
	Balance  decimal.Decimal `json:"balance" db:"balance"`
	Currency string          `json:"currency" db:"currency"`
}

// Account validates the given Account structure
func (p Account) Validate() error {
	if p.ID == "" {
		return errors.New("empty ID")
	}
	if p.Currency == "" {
		return errors.New("empty Currency")
	}

	return nil
}
