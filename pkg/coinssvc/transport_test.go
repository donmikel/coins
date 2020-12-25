package coinssvc

import (
	"context"
	"github.com/donmikel/coins/pkg/coins"
	"github.com/donmikel/coins/pkg/payment"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
	"time"
)

type mockService struct {
	onGetAvailableAccounts func(ctx context.Context) (accounts []string, err error)
	onGetAllPayments       func(ctx context.Context) (payments []payment.Payment, err error)
	onSendPayments         func(ctx context.Context, payment payment.PaymentInput) (err error)
}

func (m *mockService) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	return m.onGetAvailableAccounts(ctx)
}

func (m *mockService) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	return m.onGetAllPayments(ctx)
}

func (m *mockService) SendPayment(ctx context.Context, payment payment.PaymentInput) (err error) {
	return m.onSendPayments(ctx, payment)
}

func initTransportTest(t *testing.T) (*httptest.Server, *Client, *mockService) {
	svc := &mockService{}
	handler := makeHandler(svc)
	server := httptest.NewServer(handler)
	client, err := NewClient(ClientConfig{
		ServiceURL: server.URL,
		Timeout:    time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	return server, client, svc
}

func TestTransportGetAvailableAccounts(t *testing.T) {
	server, client, svc := initTransportTest(t)
	defer server.Close()

	testCases := []struct {
		name   string
		result []string
		err    error
	}{
		{
			name: "ok",
			result: []string{
				"bob123",
				"alice456",
			},
			err: nil,
		},
		{
			name:   "error bad request",
			result: nil,
			err:    coins.ErrBadRequest("some validation error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc.onGetAvailableAccounts = func(ctx context.Context) (accounts []string, err error) {
				return tc.result, tc.err
			}

			gotResult, gotErr := client.GetAvailableAccounts(context.Background())

			assert.Equal(t, tc.err, gotErr)
			assert.Equal(t, tc.result, gotResult)
		})
	}
}

func TestTransportGetAllPayments(t *testing.T) {
	server, client, svc := initTransportTest(t)
	defer server.Close()

	testCases := []struct {
		name    string
		result  []payment.Payment
		wantErr error
	}{
		{
			name: "ok",
			result: []payment.Payment{
				mustNewPayment(nil),
			},
			wantErr: nil,
		},
		{
			name:    "error bad request",
			result:  nil,
			wantErr: coins.ErrBadRequest("some validation error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc.onGetAllPayments = func(ctx context.Context) (payments []payment.Payment, err error) {
				return tc.result, tc.wantErr
			}

			gotResult, gotErr := client.GetAllPayments(context.Background())

			assert.Equal(t, tc.wantErr, gotErr)
			assert.Equal(t, tc.result, gotResult)
		})
	}
}

func TestTransportSendPayment(t *testing.T) {
	server, client, svc := initTransportTest(t)
	defer server.Close()

	testCases := []struct {
		name    string
		input   payment.PaymentInput
		wantErr error
	}{
		{
			name:    "ok",
			input:   mustNewPaymentInput(nil),
			wantErr: nil,
		},
		{
			name:    "error bad request",
			input:   mustNewPaymentInput(nil),
			wantErr: coins.ErrBadRequest("some validation error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotInput payment.PaymentInput
			svc.onSendPayments = func(ctx context.Context, payment payment.PaymentInput) (err error) {
				gotInput = payment
				return tc.wantErr
			}

			gotErr := client.SendPayment(context.Background(), tc.input)

			assert.Equal(t, tc.input, gotInput)
			assert.Equal(t, tc.wantErr, gotErr)
		})
	}
}
func mustNewPayment(fn func(pi payment.Payment)) payment.Payment {
	pi := payment.Payment{
		ID:          1,
		FromAccount: "bob123",
		ToAccount:   "alice456",
		Amount:      decimal.NewFromInt(100),
		Direction:   payment.Incomming,
		Dt:          nil,
	}
	if fn != nil {
		fn(pi)
	}
	return pi
}
func mustNewPaymentInput(fn func(pi payment.PaymentInput)) payment.PaymentInput {
	pi := payment.PaymentInput{
		FromAccount: "bob123",
		ToAccount:   "alice456",
		Amount:      decimal.NewFromInt(100),
		Direction:   payment.Incomming,
	}
	if fn != nil {
		fn(pi)
	}
	return pi
}
