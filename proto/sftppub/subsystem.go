package sftppub

import (
	"context"
	"errors"
	"io"

	"github.com/go-git/go-billy/v5"
	"github.com/pkg/sftp"
	"go.uber.org/zap"

	"github.com/liclac/pubd/proto/sshpub"
)

type Subsystem struct{}

func New() sshpub.Subsystem {
	return Subsystem{}
}

func (s Subsystem) Exec(ctx context.Context, L *zap.Logger, fs billy.Filesystem, c io.ReadWriteCloser) error {
	srv := sftp.NewRequestServer(c, FileSystemHandlers(L, fs))
	errC := make(chan error, 1)
	go func() {
		defer close(errC)
		if err := srv.Serve(); !errors.Is(err, io.EOF) {
			errC <- err
		}
	}()
	select {
	case err := <-errC:
		srv.Close()
		return err
	case <-ctx.Done():
		return srv.Close()
	}
}
