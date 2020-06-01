package httppub

import (
	"context"
	"net"
	"net/http"
)

// Serves HTTP requests until the context terminates, then closes the
// listener in order to shut down gracefully.
func Serve(ctx context.Context, l net.Listener, h http.Handler) error {
	srv := http.Server{Handler: h}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
