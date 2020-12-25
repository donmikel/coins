package coinssvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/donmikel/coins/pkg/coins"
	"github.com/donmikel/coins/pkg/payment"
)

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	var e *coins.ServiceError
	ok := errors.As(err, &e)
	if !ok {
		e = &coins.ServiceError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	e.Encode(w)
}

func decodeError(r *http.Response) error {
	e := &coins.ServiceError{}
	e.Decode(r)

	return e
}

type getAllPaymentsRequest struct {
}

type getAllPaymentsResponse struct {
	payments []payment.Payment
}

func encodeGetAllPaymentsRequest(ctx context.Context, r *http.Request, request interface{}) error {
	r.URL.Path = "/api/v1/payments"
	return nil
}

func decodeGetAllPaymentsResponse(ctx context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, decodeError(r)
	}
	res := getAllPaymentsResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res.payments); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return res, nil
}

func decodeGetAllPaymentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return getAllPaymentsRequest{}, nil
}

func encodeGetAllPaymentsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(getAllPaymentsResponse)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.payments); err != nil {
		return coins.ErrInternal("failed to encode JSON response: %s", err)
	}

	return nil
}

type sendPaymentRequest struct {
	input payment.PaymentInput
}

type sendPaymentResponse struct {
}

func encodeSendPaymentRequest(ctx context.Context, r *http.Request, request interface{}) error {
	req := request.(sendPaymentRequest)
	r.URL.Path = "/api/v1/payments"
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req.input); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)

	return nil
}

func decodeSendPaymentResponse(ctx context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, decodeError(r)
	}

	return sendPaymentResponse{}, nil
}

func decodeSendPaymentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var input payment.PaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return nil, coins.ErrBadRequest("failed to decode JSON request: %v", err)
	}

	return sendPaymentRequest{input: input}, nil
}

func encodeSendPaymentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusOK)

	return nil
}

type getAvailableAccountsRequest struct {
}

type getAvailableAccountsResponse struct {
	accounts []string
}

func encodeGetAvailableAccountsRequest(ctx context.Context, r *http.Request, request interface{}) error {
	r.URL.Path = "/api/v1/accounts"

	return nil
}

func decodeGetAvailableAccountsResponse(ctx context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode < 200 || r.StatusCode > 299 {
		return nil, decodeError(r)
	}
	res := getAvailableAccountsResponse{}
	if err := json.NewDecoder(r.Body).Decode(&res.accounts); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return res, nil
}

func decodeGetAvailableAccountsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return getAvailableAccountsRequest{}, nil
}

func encodeGetAvailableAccountsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(getAvailableAccountsResponse)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res.accounts); err != nil {
		return coins.ErrInternal("failed to encode JSON response: %s", err)
	}

	return nil
}
