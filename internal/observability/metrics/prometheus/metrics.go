package prometheus

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type Metric struct {
	srv *http.Server
	reg *prometheus.Registry
}

func New(port int) *Metric {
	return &Metric{
		srv: &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
		reg: prometheus.NewRegistry(),
	}
}

func (m *Metric) Start(ctx context.Context) {
	m.reg.MustRegister(
		RequestMetrics,
		RequestCount,
		CreatedPVZ,
		CreatedOrderReceipts,
		AddedProducts,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	mux := http.NewServeMux()
	mux.Handle(
		"/metrics", promhttp.HandlerFor(
			m.reg,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		),
	)

	m.srv.Handler = mux
	go func() {
		if err := m.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Prometheus server failed", zap.Error(err))
		}
	}()

	zap.L().Info("Starting prometheus server", zap.String("add", m.srv.Addr))

	<-ctx.Done()
	if err := m.srv.Shutdown(ctx); err != nil {
		zap.L().Debug("Prometheus server shutdown failed", zap.Error(err))
	}
	zap.L().Debug("Prometheus server has been stopped")
}

func ObserveRequest(d time.Duration, status int, endpoint string) {
	RequestMetrics.WithLabelValues(strconv.Itoa(status), endpoint).Observe(d.Seconds())
	RequestCount.WithLabelValues(strconv.Itoa(status), endpoint).Inc()
}

var RequestMetrics = promauto.NewSummaryVec(
	prometheus.SummaryOpts{
		Namespace:  "svc",
		Name:       "request_metrics",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"status", "endpoint"},
)

var RequestCount = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "svc",
		Name:      "request_count_total",
		Help:      "Total number of requests",
	},
	[]string{"status", "endpoint"},
)

var CreatedPVZ = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "svc",
		Name:      "created_pvz_total",
		Help:      "Total number of created PVZ",
	},
)

var CreatedOrderReceipts = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "svc",
		Name:      "created_order_receipts_total",
		Help:      "Total number of created order receptions",
	},
)

var AddedProducts = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "svc",
		Name:      "added_products_total",
		Help:      "Total number of added products",
	},
)
