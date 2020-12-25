package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/donmikel/coins/pkg/coinssvc"
	"github.com/donmikel/coins/pkg/storage"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/sync/errgroup"
)

const metricPrefix = "coins_payments"

type configuration struct {
	Port            string        `envconfig:"PORT" required:"true"`
	ReadTimeout     time.Duration `envconfig:"READ_TIMEOUT" default:"1s"`
	WriteTimeout    time.Duration `envconfig:"WRITE_TIMEOUT" default:"1s"`
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"1s"`
	AllowedOrigins  []string      `envconfig:"ALLOWED_ORIGINS"`

	PostgresAddress  string `envconfig:"POSTGRES_ADDRESS" required:"true"`
	PostgresDatabase string `envconfig:"POSTGRES_DATABASE" required:"true"`
	PostgresUser     string `envconfig:"POSTGRES_USER" required:"true"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD" required:"true"`
}

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "caller", log.DefaultCaller)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	ctx, cancel := signalContext(logger)
	defer cancel()

	level.Info(logger).Log("msg", "service is starting")
	if err := run(ctx, logger); err != nil {
		level.Error(logger).Log("msg", "service is stopped with error", "err", err)
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "service is stopped")
}

func run(ctx context.Context, logger log.Logger) error {
	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	storage, err := storage.New(storage.Config{
		PostgresAddress:  cfg.PostgresAddress,
		PostgresDatabase: cfg.PostgresDatabase,
		PostgresPassword: cfg.PostgresPassword,
		PostgresUser:     cfg.PostgresUser,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			level.Error(logger).Log("msg", "failed to close storage", "err", err)
		}
	}()

	srv, err := coinssvc.NewServer(coinssvc.ServerConfig{
		AllowedOrigins:  cfg.AllowedOrigins,
		Storage:         storage,
		Logger:          logger,
		Port:            cfg.Port,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
		MetricPrefix:    metricPrefix,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		level.Info(logger).Log("msg", "starting http server", "port", cfg.Port)
		if err := srv.Serve(ctx); err != nil {
			return fmt.Errorf("failed to serve http: %w", err)
		}

		return nil
	})

	return g.Wait()
}

// signalContext returns a context that is canceled if either SIGTERM or SIGINT signal is received.
func signalContext(logger log.Logger) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case sig := <-c:
			level.Info(logger).Log("msg", "received signal", "signal", sig)
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
