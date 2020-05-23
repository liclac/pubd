package httppub

import (
	"context"
	"net"
	"net/http"
)

type Server struct {
	Root http.FileSystem
}

// Serves HTTP requests until the context terminates, then closes the
// listener in order to shut down gracefully.
func (s Server) Serve(ctx context.Context, l net.Listener) error {
	srv := http.Server{Handler: http.FileServer(s.Root)}
	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()
	if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
