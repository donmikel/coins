package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/donmikel/coins/pkg/payment"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var errNoConnection = errors.New("no connection to database")

// Config is a storage configuration.
type Config struct {
	PostgresAddress  string
	PostgresDatabase string
	PostgresUser     string
	PostgresPassword string
	host             string
	port             string
}

func (pc *Config) validate() error {
	if pc.PostgresAddress == "" {
		return errors.New("empty PostgresAddress")
	}

	if pc.PostgresDatabase == "" {
		return errors.New("empty PostgresDatabase")
	}

	if pc.PostgresUser == "" {
		return errors.New("empty PostgresUser")
	}

	if pc.PostgresPassword == "" {
		return errors.New("empty PostgresPassword")
	}

	urlDetails := strings.Split(pc.PostgresAddress, ":")
	if len(urlDetails) != 2 {
		return errors.New("missing host or port")
	}

	pc.host = urlDetails[0]
	pc.port = urlDetails[1]

	return nil
}

func (pc *Config) getConnString() string {
	return strings.Join([]string{
		`dbname=` + pc.PostgresDatabase,
		`user=` + pc.PostgresUser,
		`password=` + pc.PostgresPassword,
		`host=` + pc.host,
		`port=` + pc.port,
		`sslmode=` + `disable`,
	}, " ")
}

// Storage is a Postgres persistent payments storage.
type Storage struct {
	db *sqlx.DB
}

func (s *Storage) getConn() (*sqlx.DB, error) {
	if s.db == nil {
		return nil, errNoConnection
	}

	return s.db, nil
}

// New creates a new storage.
func New(cfg Config) (*Storage, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	pgdb, err := sqlx.Open("postgres", cfg.getConnString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{
		db: pgdb,
	}

	return s, nil
}

// SendPayment function make a send payment in storage
func (s *Storage) SendPayment(ctx context.Context, payment payment.Payment) (err error) {
	conn, err := s.getConn()
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(ctx, `call send_payment_proc($1, $2, $3, $4)`, payment.FromAccount, payment.ToAccount, payment.Amount, payment.Direction)
	if err != nil {
		return fmt.Errorf("failed to call send_payment_proc: %w", err)
	}

	return nil
}

// GetAllPayments function return all payments
func (s *Storage) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	conn, err := s.getConn()
	if err != nil {
		return nil, err
	}

	payments = make([]payment.Payment, 0)
	err = conn.SelectContext(ctx, &payments, `select id, from_account, to_account, amount, direction, dt from payments`)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	return payments, nil
}

// GetAvailableAccounts function return all accounts available to send payment
func (s *Storage) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	conn, err := s.getConn()
	if err != nil {
		return nil, err
	}

	accounts = make([]string, 0)
	err = conn.SelectContext(ctx, &accounts, "select id from accounts")
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return accounts, nil
}

func (s *Storage) Close() error {
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return nil
}
