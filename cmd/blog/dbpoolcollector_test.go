package main

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDBPoolCollector_ExposesPoolStats(t *testing.T) {
	// pgxpool.Stat wraps an unexported *puddle.Stat with no public
	// constructor, so there's no way to fake a populated Stat directly —
	// the only practical seam is a real (but never-connecting; pgxpool
	// connects lazily on first Acquire) pool. MaxConns is set explicitly
	// so the expected value doesn't depend on the test machine's CPU
	// count (pgxpool otherwise defaults it from runtime.NumCPU()).
	config, err := pgxpool.ParseConfig("postgres://placeholder/db")
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	collector := newDBPoolCollector(pool)

	expected := `
		# HELP blog_db_pool_max_conns Maximum size of the pool.
		# TYPE blog_db_pool_max_conns gauge
		blog_db_pool_max_conns 5
		# HELP blog_db_pool_acquired_conns Number of currently acquired connections in the pool.
		# TYPE blog_db_pool_acquired_conns gauge
		blog_db_pool_acquired_conns 0
		# HELP blog_db_pool_idle_conns Number of currently idle connections in the pool.
		# TYPE blog_db_pool_idle_conns gauge
		blog_db_pool_idle_conns 0
		# HELP blog_db_pool_total_conns Total number of connections currently open (acquired + idle + constructing).
		# TYPE blog_db_pool_total_conns gauge
		blog_db_pool_total_conns 0
		# HELP blog_db_pool_acquire_count_total Cumulative count of successful connection acquires.
		# TYPE blog_db_pool_acquire_count_total counter
		blog_db_pool_acquire_count_total 0
	`

	err = testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"blog_db_pool_max_conns",
		"blog_db_pool_acquired_conns",
		"blog_db_pool_idle_conns",
		"blog_db_pool_total_conns",
		"blog_db_pool_acquire_count_total",
	)
	if err != nil {
		t.Fatal(err)
	}
}
