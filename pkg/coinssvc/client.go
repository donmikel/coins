package coinssvc

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/donmikel/coins/pkg/payment"
	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

// ClientConfig is an Client configuration.
type ClientConfig struct {
	ServiceURL string
	Timeout    time.Duration
}

func (cfg ClientConfig) validate() error {
	if cfg.ServiceURL == "" {
		return errors.New("must provide ServiceURL")
	}
	if cfg.Timeout <= 0 {
		return errors.New("invalid Timeout")
	}

	return nil
}

var _ Service = (*Client)(nil)

// Client is a filters service client.
type Client struct {
	getAllPaymentsEndpoint       endpoint.Endpoint
	sendPaymentEndpoint          endpoint.Endpoint
	getAvailableAccountsEndpoint endpoint.Endpoint
}

// NewClient creates a new client.
func NewClient(cfg ClientConfig) (*Client, error) {
	err := cfg.validate()
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(cfg.ServiceURL)
	if err != nil {
		return nil, err
	}

	options := []kithttp.ClientOption{
		kithttp.SetClient(&http.Client{
			Timeout: cfg.Timeout,
		}),
	}

	c := &Client{
		getAllPaymentsEndpoint: kithttp.NewClient(
			http.MethodGet,
			baseURL,
			encodeGetAllPaymentsRequest,
			decodeGetAllPaymentsResponse,
			options...,
		).Endpoint(),
		sendPaymentEndpoint: kithttp.NewClient(
			http.MethodPost,
			baseURL,
			encodeSendPaymentRequest,
			decodeSendPaymentResponse,
			options...,
		).Endpoint(),
		getAvailableAccountsEndpoint: kithttp.NewClient(
			http.MethodGet,
			baseURL,
			encodeGetAvailableAccountsRequest,
			decodeGetAvailableAccountsResponse,
			options...,
		).Endpoint(),
	}

	return c, nil
}

// GetAllPayments get list of all payments.
func (c *Client) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	response, err := c.getAllPaymentsEndpoint(ctx, getAllPaymentsRequest{})
	if err != nil {
		return nil, err
	}

	return response.(getAllPaymentsResponse).payments, nil
}

// SendPayment send payment to user.
func (c *Client) SendPayment(ctx context.Context, payment payment.PaymentInput) (err error) {
	_, err = c.sendPaymentEndpoint(ctx, sendPaymentRequest{input: payment})
	if err != nil {
		return err
	}

	return nil
}

// GetAvailableAccounts get available account to send money.
func (c *Client) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	response, err := c.getAvailableAccountsEndpoint(ctx, getAvailableAccountsRequest{})
	if err != nil {
		return nil, err
	}

	return response.(getAvailableAccountsResponse).accounts, nil
}
