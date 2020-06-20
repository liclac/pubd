package sftppub

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/pkg/sftp"
	"go.uber.org/zap"

	"github.com/liclac/pubd/proto/sshpub"
)

type Subsystem struct {
	FS billy.Filesystem
}

func New(fs billy.Filesystem) sshpub.Subsystem {
	return Subsystem{FS: fs}
}

func (s Subsystem) Exec(ctx context.Context, L *zap.Logger, c io.ReadWriteCloser) error {
	srv := sftp.NewRequestServer(c, NewHandler(L, s.FS).Handlers())
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

type Handler struct {
	L  *zap.Logger
	FS billy.Filesystem
}

func NewHandler(L *zap.Logger, fs billy.Filesystem) Handler {
	return Handler{L, fs}
}

func (h Handler) Handlers() sftp.Handlers {
	return sftp.Handlers{FileGet: h, FilePut: h, FileCmd: h, FileList: h}
}

func (h Handler) handleErr(req *sftp.Request, err error) error {
	h.L.Error("Request Error", zap.String("method", req.Method),
		zap.String("path", req.Filepath), zap.Error(err))
	if os.IsNotExist(err) {
		return sftp.ErrSshFxNoSuchFile
	} else if os.IsPermission(err) {
		return sftp.ErrSshFxPermissionDenied
	}
	return sftp.ErrSSHFxFailure
}

func (h Handler) Fileread(req *sftp.Request) (io.ReaderAt, error) {
	switch req.Method {
	case "Get":
		f, err := h.FS.Open(req.Filepath)
		if err != nil {
			return nil, h.handleErr(req, err)
		}
		return f, nil
	default:
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

func (h Handler) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	switch req.Method {
	case "Put", "Open":
		return nil, sftp.ErrSshFxPermissionDenied
	default:
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

func (h Handler) Filecmd(req *sftp.Request) error {
	switch req.Method {
	case "Setstat", "Rename", "Rmdir", "Mkdir", "Link", "Symlink", "Remove":
		return sftp.ErrSshFxPermissionDenied
	default:
		return sftp.ErrSshFxOpUnsupported
	}
}

func (h Handler) Filelist(req *sftp.Request) (sftp.ListerAt, error) {
	switch req.Method {
	case "List":
		infos, err := h.FS.ReadDir(req.Filepath)
		if err != nil {
			return nil, h.handleErr(req, err)
		}
		return listerAt(infos), nil
	case "Stat":
		info, err := h.FS.Stat(req.Filepath)
		if err != nil {
			return nil, h.handleErr(req, err)
		}
		return listerAt{info}, nil
	case "Readlink":
		info, err := h.FS.Lstat(req.Filepath)
		if err != nil {
			return nil, h.handleErr(req, err)
		}
		return listerAt{info}, nil
	default:
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

type listerAt []os.FileInfo

func (f listerAt) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n := copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}
