package sshpub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"github.com/liclac/pubd"
)

var _ pubd.Server = Server{}

// A subsystem offered by the SSH session's 'subsystem' command.
type Subsystem interface {
	Exec(context.Context, *zap.Logger, io.ReadWriteCloser) error
}

type Server struct {
	L       *zap.Logger
	HostKey ssh.Signer

	// Silence warnings about unsupported subsystems by explicitly setting them to nil.
	Subsystems map[string]Subsystem
}

func New(L *zap.Logger, hostKey ssh.Signer) Server {
	return Server{L: L, HostKey: hostKey}
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

	var g sync.WaitGroup
	defer g.Wait()
	for {
		select {
		case newChan, ok := <-newChanC:
			if !ok {
				return // Connection closed
			}
			typ := newChan.ChannelType()
			switch typ {
			case "session":
				ch, reqC, err := newChan.Accept()
				if err != nil {
					L.Warn("Couldn't accept chanel", zap.String("type", typ))
					break
				}
				sessL := L.Named("sess")
				sessL.Debug("Session started")

				g.Add(1)
				go func() {
					defer g.Done()
					s.ServeSession(ctx, sessL, ch, reqC)
					if err := ch.Close(); err != nil && !errors.Is(err, io.EOF) {
						sessL.Debug("Error closing session", zap.Error(err))
					}
					sessL.Debug("Session closed")
				}()
			default:
				L.Warn("Unknown channel", zap.String("type", typ))
				if err := newChan.Reject(ssh.UnknownChannelType, typ); err != nil {
					L.Warn("Error rejecting channel", zap.String("type", typ), zap.Error(err))
				}
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

func (s *Server) ServeSession(ctx context.Context, L *zap.Logger, ch ssh.Channel, reqC <-chan *ssh.Request) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-reqC:
			if !ok {
				return // Connection closed
			}
			switch req.Type {
			// | RFC 4254: SSH_MSG_CHANNEL_REQUEST - "subsystem"
			// | string    subsystem name
			//
			// The client is requesting that a subsystem takes over the connection.
			// After a successful subsystem/shell/exec, the channel is closed.
			case "subsystem":
				subL := L.Named("sub")

				name, _, ok := DecodeString(req.Payload)
				if !ok {
					subL.Warn("Malformed request", zap.Binary("payload", req.Payload))
					lazyReply(L, req, false)
					break
				}
				// Distinguish between unknown subsystems and disabled (nil) ones.
				sys, ok := s.Subsystems[name]
				if !ok {
					subL.Warn("Unsupported subsystem", zap.String("name", name))
					lazyReply(L, req, false)
					break
				}
				if sys == nil {
					subL.Debug("Disabled subsystem", zap.String("name", name))
					lazyReply(L, req, false)
					break
				}

				// The subsystem is supported, let it take the wheel.
				L = L.Named(name)
				L.Debug("Starting")
				if req.WantReply {
					if err := req.Reply(true, nil); err != nil {
						L.Error("Reply error", zap.Error(err))
						return
					}
				}
				if err := sys.Exec(ctx, L, ch); err != nil {
					L.Error("Error", zap.Error(err))
				}
				L.Debug("Finished")
				return

			// If we wanted to offer a shell, this would be the place to do it.
			// For now, just print a nicer error message and return exit code 255.
			case "shell", "exec":
				if req.WantReply {
					if err := req.Reply(true, nil); err != nil {
						L.Error("Error accepting command", zap.String("type", req.Type), zap.Error(err))
						return
					}
				}
				fmt.Fprintln(ch, "// Shell access is not allowed.")
				ch.SendRequest("exit-status", false, EncodeUint32(255))
				return

			// Discard shell/exec's supporting commands, else the openssh client prints errors.
			case "pty-req", "env", "window-change":
				L.Debug("Discarding request", zap.String("type", req.Type))
				lazyReply(L, req, true)

			// Known unsupported commands.
			case "x11-req", "xon-xoff", "signal":
				L.Debug("Unsupported request", zap.String("type", req.Type))
				lazyReply(L, req, false)

			// ???
			default:
				L.Warn("Unknown request", zap.String("type", req.Type))
				lazyReply(L, req, false)
			}
		}
	}
}
