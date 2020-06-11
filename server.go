package pubd

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
)

type Server interface {
	// Accepts connections on l until ctx expires, then gracefully shuts down.
	Serve(ctx context.Context, l net.Listener) error
}

// Wraps a Serve closure in a Server.
type ServerFunc func(ctx context.Context, l net.Listener) error

func (fn ServerFunc) Serve(ctx context.Context, l net.Listener) error {
	return fn(ctx, l)
}

// Runs an instance of srv for each listener. The context passed to each srv is a child
// of ctx, which is cancelled when the first instance returns.
func Serve(ctx context.Context, listeners []net.Listener, srv Server) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, l := range listeners {
		l := l
		g.Go(func() error { return srv.Serve(ctx, l) })
	}
	return g.Wait()
}

// Shorthand for calling Serve(ctx, Listen(addr), srv).
func ListenAndServe(ctx context.Context, addr string, srv Server) error {
	listeners, err := Listen(addr)
	if err != nil {
		return err
	}
	return Serve(ctx, listeners, srv)
}
