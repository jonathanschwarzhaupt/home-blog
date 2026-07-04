package main

import (
	"context"
	"testing"
	"time"

	"github.com/jonathanschwarzhaupt/my-blog/internal/assert"
)

func TestServe_GracefulShutdown(t *testing.T) {
	app := newTestApplication()
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- serve(ctx, app, "127.0.0.1:0")
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		assert.Nil(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("serve did not shut down within the expected time")
	}
}
