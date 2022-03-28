package kustoutils

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// we want a series with run status 0-fail, 1-partial, 2-success
var (
	deltaSuccesses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "delat_kusto_success",
			Help: "Number of successful delta kusto runs",
		},
		[]string{"cluster"},
	)
	deltaFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "delat_kusto_failures",
			Help: "Number of failed delta kusto runs",
		},
		[]string{"cluster"},
	)
	deltaDurationsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "delta_durations_histogram_minutes",
		Help:    "delta kusto run duration distributions.",
		Buckets: prometheus.LinearBuckets(0, 1, 20),
	},
		[]string{"cluster"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(deltaSuccesses, deltaFailures)
	addFailure("test")
	addSuccess("test")
	addDuration("test", 3.2)

}

func addFailure(cluster string) {
	deltaFailures.WithLabelValues(cluster).Inc()
}

func addSuccess(cluster string) {
	deltaSuccesses.WithLabelValues(cluster).Inc()
}

func addDuration(cluster string, duration float64) {
	deltaDurationsHistogram.WithLabelValues(cluster).Observe(duration)
}
