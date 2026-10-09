package telemetry

import (
	"context"
	"database/sql"
	"time"

	"github.com/octoplorer/octopulse/internal/store"
	"github.com/prometheus/client_golang/prometheus"
)

// ObserveStore uses snapshots so database handles stay private to the store.
func (m *Metrics) ObserveStore(st *store.Store) {
	pools := map[string]func() sql.DBStats{"write": func() sql.DBStats { return st.Stats().Write }}
	if st.Driver() == "sqlite" {
		pools["read"] = func() sql.DBStats { return st.Stats().Read }
	}
	for pool, snapshot := range pools {
		labels := prometheus.Labels{"pool": pool}
		for _, metric := range []struct {
			name, help string
			value      func(sql.DBStats) float64
		}{
			{
				name: "open_connections",
				help: "Open database connections.",
				value: func(s sql.DBStats) float64 {
					return float64(s.OpenConnections)
				},
			},
			{
				name: "in_use_connections",
				help: "Database connections currently in use.",
				value: func(s sql.DBStats) float64 {
					return float64(s.InUse)
				},
			},
			{
				name: "idle_connections",
				help: "Idle database connections.",
				value: func(s sql.DBStats) float64 {
					return float64(s.Idle)
				},
			},
			{
				name: "max_open_connections",
				help: "Configured maximum open database connections.",
				value: func(s sql.DBStats) float64 {
					return float64(s.MaxOpenConnections)
				},
			},
		} {
			m.registry.MustRegister(
				prometheus.NewGaugeFunc(
					prometheus.GaugeOpts{
						Name:        "octopulse_db_" + metric.name,
						Help:        metric.help,
						ConstLabels: labels,
					},
					func() float64 {
						return metric.value(snapshot())
					},
				),
			)
		}
		m.registry.MustRegister(
			prometheus.NewCounterFunc(
				prometheus.CounterOpts{
					Name:        "octopulse_db_waits_total",
					Help:        "Waits for an available database connection.",
					ConstLabels: labels,
				},
				func() float64 {
					return float64(snapshot().WaitCount)
				},
			),
			prometheus.NewCounterFunc(
				prometheus.CounterOpts{
					Name:        "octopulse_db_wait_duration_seconds_total",
					Help:        "Time spent waiting for available database connections.",
					ConstLabels: labels,
				},
				func() float64 { return snapshot().WaitDuration.Seconds() },
			),
		)
	}
	m.registry.MustRegister(&backlogCollector{store: st,
		pending: prometheus.NewDesc(
			"octopulse_delivery_pending",
			"Pending notification deliveries.",
			nil,
			nil,
		),
		age: prometheus.NewDesc(
			"octopulse_delivery_oldest_due_age_seconds",
			"Age of the oldest overdue pending delivery.",
			nil,
			nil,
		),
		success: prometheus.NewDesc(
			"octopulse_delivery_backlog_scrape_success",
			"Whether the delivery backlog could be read.",
			nil,
			nil,
		),
	})
}

type backlogCollector struct {
	store                 *store.Store
	pending, age, success *prometheus.Desc
}

func (c *backlogCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.pending
	ch <- c.age
	ch <- c.success
}
func (c *backlogCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pending, due, err := c.store.DeliveryBacklog(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.success, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.pending, prometheus.GaugeValue, float64(pending))
	age := float64(0)
	if pending > 0 {
		age = max(0, float64(time.Now().UnixMilli()-due)/1000)
	}
	ch <- prometheus.MustNewConstMetric(c.age, prometheus.GaugeValue, age)
}
