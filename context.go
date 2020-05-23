package pubd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Returns a context which is cancelled upon receiving SIGINT or SIGTERM.
func WithSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(sigC)
		select {
		case <-ctx.Done():
		case <-sigC:
			cancel()
		}
	}()
	return ctx
}
