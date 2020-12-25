package coinssvc

import (
	"context"
	"strconv"
	"time"

	"github.com/donmikel/coins/pkg/payment"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// InstrumentingMiddleware wraps Service and records metrics.
type InstrumentingMiddleware struct {
	svc       Service
	histogram metrics.Histogram
}

func NewInstrumentingMiddleware(svc Service, prefix string) *InstrumentingMiddleware {
	return &InstrumentingMiddleware{
		svc: svc,
		histogram: kitprometheus.NewHistogramFrom(
			prometheus.HistogramOpts{
				Name:    prefix + "_requests",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 14),
			},
			[]string{"method", "error"},
		),
	}
}

func (mw *InstrumentingMiddleware) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	defer mw.record(time.Now(), "GetAllPayments", &err)
	return mw.svc.GetAllPayments(ctx)
}

func (mw *InstrumentingMiddleware) SendPayment(ctx context.Context, payment payment.PaymentInput) (err error) {
	defer mw.record(time.Now(), "SendPayment", &err)
	return mw.svc.SendPayment(ctx, payment)
}

func (mw *InstrumentingMiddleware) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	defer mw.record(time.Now(), "GetAvailableAccounts", &err)
	return mw.svc.GetAvailableAccounts(ctx)
}

func (mw *InstrumentingMiddleware) record(beginTime time.Time, method string, err *error) {
	labels := []string{"method", method, "error", strconv.FormatBool(*err != nil)}
	mw.histogram.With(labels...).Observe(time.Since(beginTime).Seconds())
}

// LoggingMiddleware wraps Service and logs errors.
type LoggingMiddleware struct {
	svc    Service
	logger log.Logger
}

func NewLoggingMiddleware(svc Service, logger log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		svc:    svc,
		logger: logger,
	}
}

func (mw *LoggingMiddleware) GetAllPayments(ctx context.Context) (payments []payment.Payment, err error) {
	defer mw.log(time.Now(), "GetAllPayments", &err)
	return mw.svc.GetAllPayments(ctx)
}

func (mw *LoggingMiddleware) SendPayment(ctx context.Context, payment payment.PaymentInput) (err error) {
	defer mw.log(time.Now(), "SendPayment", &err)
	return mw.svc.SendPayment(ctx, payment)
}

func (mw *LoggingMiddleware) GetAvailableAccounts(ctx context.Context) (accounts []string, err error) {
	defer mw.log(time.Now(), "GetAvailableAccounts", &err)
	return mw.svc.GetAvailableAccounts(ctx)
}

func (mw *LoggingMiddleware) log(beginTime time.Time, method string, err *error) {
	if *err != nil {
		level.Error(mw.logger).Log("method", method, "err", *err, "took", time.Since(beginTime))
	}
}
