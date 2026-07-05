package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

func serve(ctx context.Context, app *application, addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error, 1)

	go func() {
		<-ctx.Done()

		app.logger.Info("shutting down server", "addr", srv.Addr)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(shutdownCtx)
	}()

	app.logger.Info("starting server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)
	return nil
}

// serveMetrics mirrors serve()'s graceful-shutdown shape for the separate
// metrics listener — a distinct *http.Server on its own port, shut down on
// the same ctx as the main server, but with no dependency on *application
// since it only ever serves the Prometheus handler.
func serveMetrics(ctx context.Context, logger *slog.Logger, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	shutdownError := make(chan error, 1)

	go func() {
		<-ctx.Done()

		logger.Info("shutting down metrics server", "addr", srv.Addr)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(shutdownCtx)
	}()

	logger.Info("starting metrics server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	if err := <-shutdownError; err != nil {
		return err
	}

	logger.Info("stopped metrics server", "addr", srv.Addr)
	return nil
}
