package sshpub

import (
	"context"
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"github.com/liclac/pubd"
)

var _ pubd.Server = Server{}

type Server struct {
	L       *zap.Logger
	HostKey ssh.Signer
}

func New(L *zap.Logger, hostKey ssh.Signer) Server {
	srv := Server{
		L:       L,
		HostKey: hostKey,
	}
	return srv
}

func (s Server) Serve(ctx context.Context, l net.Listener) error {
	cfg := ssh.ServerConfig{NoClientAuth: true}
	cfg.AddHostKey(s.HostKey)

	// Make sure the listener doesn't leak if we error out.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Graceful shutdown begins by not accepting new connections.
	go func() {
		<-ctx.Done()
		l.Close()
	}()

	// Then waiting for existing ones to disconnect.
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		// Accept a new connection!
		nConn, err := l.Accept()
		if err != nil {
			// Failure to accept a new connection is fatal, but not necessarily an error.
			select {
			case <-ctx.Done():
				// If the context is cancelled, l will be closed and Accept() will error.
				return nil
			default:
				return fmt.Errorf("couldn't accept connection: %w", err)
			}
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			s.ServeConn(ctx, nConn, cfg)
		}()
	}
}

func (s *Server) ServeConn(ctx context.Context, nConn net.Conn, cfg ssh.ServerConfig) {
	L := s.L.With(zap.Stringer("addr", nConn.RemoteAddr()))

	// Log authentication attempts.
	authL := L.Named("auth")
	cfg.AuthLogCallback = func(meta ssh.ConnMetadata, method string, err error) {
		authL := authL.With(zap.String("user", meta.User()), zap.String("method", method))
		if err != nil {
			authL.Warn("Attempt failed", zap.Error(err))
		} else {
			authL.Debug("Accepted")
		}
	}

	// SSH handshake!
	sConn, newChanC, reqC, err := ssh.NewServerConn(nConn, &cfg)
	if err != nil {
		defer nConn.Close()
		if authErr, ok := err.(*ssh.ServerAuthError); ok {
			authL.Warn("Authentication failed", zap.Error(authErr))
			return
		}
		L.Error("Handshake failed", zap.Error(err))
		return
	}
	L = L.With(zap.String("user", sConn.User()))
	defer sConn.Close()

	connL := L.Named("conn")
	connL.Info("Connected")
	defer connL.Info("Disconnected")
	for {
		select {
		case newChan, ok := <-newChanC:
			if !ok {
				return // Connection closed
			}
			typ := newChan.ChannelType()
			L.Warn("Unknown channel requested", zap.String("type", typ))
			if err := newChan.Reject(ssh.UnknownChannelType, typ); err != nil {
				L.Warn("Error rejecting channel", zap.String("type", typ), zap.Error(err))
			}
		case req, ok := <-reqC:
			if !ok {
				return // Connection closed
			}
			L.Debug("Discarding out-of-session request", zap.String("type", req.Type))
		case <-ctx.Done():
			L.Debug("Context expired", zap.Error(ctx.Err()))
			return
		}
	}
}
