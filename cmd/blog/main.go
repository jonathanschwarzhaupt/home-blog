package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jonathanschwarzhaupt/my-blog/internal/database"
	"github.com/jonathanschwarzhaupt/my-blog/internal/models"
	"github.com/jonathanschwarzhaupt/my-blog/internal/vcs"
)

type application struct {
	logger  *slog.Logger
	db      database.Querier
	limiter *rateLimiter
}

func main() {
	opts := parseOptions()

	if opts.displayVersion {
		fmt.Println(vcs.Version())
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := models.OpenPool(ctx, opts.dbDSN, models.PoolConfig{
		MaxConns:        int32(opts.dbMaxConns),
		MinConns:        int32(opts.dbMinConns),
		MaxConnLifetime: opts.dbMaxConnLife,
		MaxConnIdleTime: opts.dbMaxIdleTime,
	})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	limiter := newRateLimiter(opts.limiterRPS, opts.limiterBurst, opts.limiterEnabled)
	limiter.startCleanup(time.Minute, 3*time.Minute)

	app := &application{
		logger:  logger,
		db:      database.New(pool),
		limiter: limiter,
	}

	if err := serve(ctx, app, opts.addr); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
