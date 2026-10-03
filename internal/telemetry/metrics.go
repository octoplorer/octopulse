// Package telemetry records bounded, credential-free operational diagnostics.
package telemetry

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry      *prometheus.Registry
	requests      *prometheus.CounterVec
	duration      *prometheus.HistogramVec
	inFlight      prometheus.Gauge
	errors        *prometheus.CounterVec
	deliveries    *prometheus.CounterVec
	probes        *prometheus.CounterVec
	probeDuration *prometheus.HistogramVec
	queueWait     prometheus.Histogram
}

func New() *Metrics {
	m := &Metrics{
		probes:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "octopulse_probe_attempts_total", Help: "Probe attempts by protocol and success."}, []string{"type", "success"}),
		probeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "octopulse_probe_attempt_duration_seconds", Help: "Probe attempt duration by protocol.", Buckets: []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60}}, []string{"type"}),
		queueWait:     prometheus.NewHistogram(prometheus.HistogramOpts{Name: "octopulse_probe_queue_wait_seconds", Help: "Accepted probe round queue wait.", Buckets: []float64{.001, .01, .1, 1, 5, 15, 30, 60}}),
		registry:      prometheus.NewRegistry(),
		requests:      prometheus.NewCounterVec(prometheus.CounterOpts{Name: "octopulse_http_requests_total", Help: "Completed HTTP requests by registered route, method and status."}, []string{"route", "method", "status"}),
		duration:      prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "octopulse_http_request_duration_seconds", Help: "HTTP request duration by registered route.", Buckets: prometheus.DefBuckets}, []string{"route", "method"}),
		inFlight:      prometheus.NewGauge(prometheus.GaugeOpts{Name: "octopulse_http_requests_in_flight", Help: "HTTP requests currently being handled."}),
		errors:        prometheus.NewCounterVec(prometheus.CounterOpts{Name: "octopulse_operation_errors_total", Help: "Background operation failures."}, []string{"operation"}),
		deliveries:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "octopulse_delivery_completions_total", Help: "Persisted delivery outcomes, including scheduled retries."}, []string{"state"}),
	}
	m.registry.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), m.requests, m.duration, m.inFlight, m.errors, m.deliveries, m.probes, m.probeDuration, m.queueWait)
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{MaxRequestsInFlight: 4, Timeout: 5 * time.Second})
}

func (m *Metrics) Gauge(name, help string, read func() float64) {
	m.registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "octopulse_" + name, Help: help}, read))
}

func (m *Metrics) BeginRequest() { m.inFlight.Inc() }
func (m *Metrics) EndRequest(route, method string, status int, elapsed time.Duration) {
	m.inFlight.Dec()
	if route == "" {
		route = "unmatched"
	}
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE":
	default:
		method = "OTHER"
	}
	m.requests.WithLabelValues(route, method, strconv.Itoa(status)).Inc()
	m.duration.WithLabelValues(route, method).Observe(elapsed.Seconds())
}

func (m *Metrics) OperationError(operation string) { m.errors.WithLabelValues(operation).Inc() }
func (m *Metrics) Delivery(state string)           { m.deliveries.WithLabelValues(state).Inc() }

func (m *Metrics) Probe(kind string, success bool, duration time.Duration) {
	switch kind {
	case "http", "tcp", "dns", "certificate":
	default:
		kind = "other"
	}
	m.probes.WithLabelValues(kind, strconv.FormatBool(success)).Inc()
	m.probeDuration.WithLabelValues(kind).Observe(duration.Seconds())
}
func (m *Metrics) QueueWait(delay time.Duration) { m.queueWait.Observe(delay.Seconds()) }
