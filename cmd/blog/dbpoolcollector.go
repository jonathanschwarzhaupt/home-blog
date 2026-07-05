package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// dbPoolCollector exposes pgxpool's own Stat() as Prometheus metrics —
// implements prometheus.Collector directly rather than a fixed set of
// gauges updated on a timer, so every scrape reads a fresh Stat().
type dbPoolCollector struct {
	pool *pgxpool.Pool

	maxConns      *prometheus.Desc
	acquiredConns *prometheus.Desc
	idleConns     *prometheus.Desc
	totalConns    *prometheus.Desc
	acquireCount  *prometheus.Desc
}

func newDBPoolCollector(pool *pgxpool.Pool) *dbPoolCollector {
	return &dbPoolCollector{
		pool: pool,
		maxConns: prometheus.NewDesc(
			"blog_db_pool_max_conns", "Maximum size of the pool.", nil, nil),
		acquiredConns: prometheus.NewDesc(
			"blog_db_pool_acquired_conns", "Number of currently acquired connections in the pool.", nil, nil),
		idleConns: prometheus.NewDesc(
			"blog_db_pool_idle_conns", "Number of currently idle connections in the pool.", nil, nil),
		totalConns: prometheus.NewDesc(
			"blog_db_pool_total_conns", "Total number of connections currently open (acquired + idle + constructing).", nil, nil),
		acquireCount: prometheus.NewDesc(
			"blog_db_pool_acquire_count_total", "Cumulative count of successful connection acquires.", nil, nil),
	}
}

func (c *dbPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxConns
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.acquireCount
}

func (c *dbPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(stat.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(stat.AcquireCount()))
}
