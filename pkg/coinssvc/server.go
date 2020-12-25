package coinssvc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/donmikel/coins/pkg/payment"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ServerConfig is a server configuration.
type ServerConfig struct {
	AllowedOrigins  []string
	Logger          log.Logger
	Storage         Storage
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MetricPrefix    string
}

// Storage is a persistent accounts data storage.
type Storage interface {
	GetAllPayments(ctx context.Context) (payments []payment.Payment, err error)
	SendPayment(ctx context.Context, payment payment.Payment) (err error)
	GetAvailableAccounts(ctx context.Context) (accounts []string, err error)
}

// Server is a accounts service server.
type Server struct {
	cfg *ServerConfig
	srv *http.Server
}

// NewServer creates a new server.
func NewServer(cfg ServerConfig) (*Server, error) {
	var svc Service
	svc = newService(cfg.Logger, cfg.Storage)
	svc = NewLoggingMiddleware(svc, cfg.Logger)
	svc = NewInstrumentingMiddleware(svc, cfg.MetricPrefix)

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Handle("/api/v1/", makeHandler(svc))

	var handler http.Handler
	if len(cfg.AllowedOrigins) == 0 {
		handler = router
	} else {
		handler = handlers.CORS(
			handlers.AllowedMethods([]string{
				http.MethodGet,
				http.MethodPost,
				http.MethodPatch,
				http.MethodDelete,
			}),
			handlers.AllowedHeaders([]string{"Content-Type"}),
			handlers.AllowedOrigins(cfg.AllowedOrigins),
		)(router)
	}

	srv := &http.Server{
		Handler:      handler,
		Addr:         ":" + cfg.Port,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	s := &Server{
		cfg: &cfg,
		srv: srv,
	}
	return s, nil
}

func makeHandler(svc Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	router := mux.NewRouter()

	router.Path("/api/v1/payments").Methods(http.MethodGet).Handler(kithttp.NewServer(
		makeGetAllPaymentsEndpoint(svc),
		decodeGetAllPaymentsRequest,
		encodeGetAllPaymentsResponse,
		opts...,
	))

	router.Path("/api/v1/payments").Methods(http.MethodPost).Handler(kithttp.NewServer(
		makeSendPaymentEndpoint(svc),
		decodeSendPaymentRequest,
		encodeSendPaymentResponse,
		opts...,
	))

	router.Path("/api/v1/accounts").Methods(http.MethodGet).Handler(kithttp.NewServer(
		makeGetAvailableAccountsEndpoint(svc),
		decodeGetAvailableAccountsRequest,
		encodeGetAvailableAccountsResponse,
		opts...,
	))

	return router
}

func makeGetAllPaymentsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		payments, err := svc.GetAllPayments(ctx)
		return getAllPaymentsResponse{payments: payments}, err
	}
}

func makeSendPaymentEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(sendPaymentRequest)
		err := svc.SendPayment(ctx, req.input)
		return sendPaymentResponse{}, err
	}
}

func makeGetAvailableAccountsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		accounts, err := svc.GetAvailableAccounts(ctx)
		return getAvailableAccountsResponse{accounts: accounts}, err
	}
}

// Serve starts HTTP server and stops it when the provided context is canceled.
func (s *Server) Serve(ctx context.Context) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.srv.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err

	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := s.srv.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
		return nil
	}
}
