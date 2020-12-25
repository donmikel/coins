package storage

import (
	"context"
	"fmt"
	"github.com/donmikel/coins/pkg/payment"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

// Environment variables used to connect to a test PostgresSQL database.
// These variables must be present in the environment for the PostgresSQL-dependent
// tests to run, otherwise they will be skipped.
const (
	testEnvAddress  = "TEST_POSTGRES_ADDRESS"
	testEnvDatabase = "TEST_POSTGRES_DATABASE"
	testEnvPort     = "TEST_POSTGRES_PORT"
	testEnvPass     = "TEST_POSTGRES_PASSWORD"
	testEnvUser     = "TEST_POSTGRES_USER"
)

func getTestStorage(tb testing.TB) (*Storage, func()) {
	address := os.Getenv(testEnvAddress)
	if address == "" {
		tb.Skipf("skipping test: environment variable not found: %s", testEnvAddress)
	}
	database := os.Getenv(testEnvDatabase)
	if database == "" {
		tb.Skipf("skipping test: environment variable not found: %s", testEnvDatabase)
	}
	port := os.Getenv(testEnvPort)
	if port == "" {
		tb.Skipf("skipping test: environment variable not found: %s", testEnvPort)
	}
	pass := os.Getenv(testEnvPass)
	if pass == "" {
		tb.Skipf("skipping test: environment variable not found: %s", testEnvPass)
	}
	user := os.Getenv(testEnvUser)
	if user == "" {
		tb.Skipf("skipping test: environment variable not found: %s", testEnvUser)
	}
	cs := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s port=%s", user, database, pass, address, port)
	tb.Log(cs)
	db, err := sqlx.Connect("postgres", cs)
	if err != nil {
		tb.Fatalf("failed to connect to postgres: %s, pass: %s", err, pass)
	}
	defer func() {
		if err := db.Close(); err != nil {
			tb.Fatalf("failed to close storage: %s", err)
		}
	}()

	//Clear test data

	_, err = db.Exec("TRUNCATE TABLE accounts; TRUNCATE TABLE payments; ")
	if err != nil {
		tb.Fatalf("failed to truncate table: %s", err)
	}

	//Add test accounts
	_, err = db.Exec("INSERT INTO accounts VALUES ('bob123', 100, 'USD'), ('alice456', 0.01, 'USD');")
	if err != nil {
		tb.Fatalf("failed to insert test dats to postgres: %s", err)
	}

	// Create a test storage.
	s, err := New(Config{
		PostgresAddress:  strings.Join([]string{address, port}, ":"),
		PostgresUser:     user,
		PostgresPassword: pass,
		PostgresDatabase: database,
	})
	if err != nil {
		tb.Fatalf("failed to create test storage: %s", err)
	}

	teardown := func() {
		if err := s.Close(); err != nil {
			tb.Fatalf("failed to close test storage: %s", err)
		}
	}
	return s, teardown
}

func TestAccounts(t *testing.T) {
	s, teardown := getTestStorage(t)
	defer teardown()

	wantAccounts := []string{
		"bob123",
		"alice456",
	}

	result, err := s.GetAvailableAccounts(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	assert.ElementsMatch(t, wantAccounts, result)
}

func TestPayments(t *testing.T) {
	s, teardown := getTestStorage(t)
	defer teardown()

	wantPayments := []payment.Payment{
		{FromAccount: "alice456",
			Amount:    decimal.NewFromFloat(0.01),
			ToAccount: "bob123",
			Direction: 0,
		},
		{FromAccount: "bob123",
			Amount:    decimal.NewFromFloat(100),
			ToAccount: "alice456",
			Direction: 1,
		},
	}
	t.Log(wantPayments)
	for _, wantPayment := range wantPayments {
		err := s.SendPayment(context.Background(), wantPayment)
		t.Log(wantPayment)
		if err != nil {
			t.Fatal(err)
		}
		var resultPayment payment.Payment
		s.db.Get(&resultPayment, "SELECT * FROM payments WHERE from_account = $1", wantPayment.FromAccount)

		wf, _ := wantPayment.Amount.Float64()
		rf, _ := resultPayment.Amount.Float64()

		assert.Equal(t, wantPayment.FromAccount, resultPayment.FromAccount)
		assert.Equal(t, wantPayment.ToAccount, resultPayment.ToAccount)
		assert.Equal(t, wf, rf)
		assert.Equal(t, wantPayment.Direction, resultPayment.Direction)
	}

	allPayments, err := s.GetAllPayments(context.Background())
	t.Log(len(allPayments))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(wantPayments)
	t.Log(allPayments)

	for i := 0; i < len(allPayments); i++ {
		wantPayment := wantPayments[i]
		resultPayment := allPayments[i]
		wf, _ := wantPayment.Amount.Float64()
		rf, _ := resultPayment.Amount.Float64()

		assert.Equal(t, wantPayment.FromAccount, resultPayment.FromAccount)
		assert.Equal(t, wantPayment.ToAccount, resultPayment.ToAccount)
		assert.Equal(t, wf, rf)
		assert.Equal(t, wantPayment.Direction, resultPayment.Direction)
	}
}

func mustNewPayment(fn func(c payment.Payment)) payment.Payment {
	c := payment.Payment{
		FromAccount: "bob123",
		Amount:      decimal.NewFromInt(100),
		ToAccount:   "alice456",
		Direction:   1,
	}

	if fn != nil {
		fn(c)
	}
	return c
}
