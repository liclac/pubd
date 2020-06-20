package sftppub

import (
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
)

type fileSystemHandlers struct {
	L  *zap.Logger
	FS billy.Filesystem
}

func FileSystemHandlers(L *zap.Logger, fs billy.Filesystem) sftp.Handlers {
	handlers := fileSystemHandlers{L, fs}
	return sftp.Handlers{
		FileGet:  handlers,
		FilePut:  handlers,
		FileCmd:  handlers,
		FileList: handlers,
	}
}

func (h fileSystemHandlers) handleErr(req *sftp.Request, err error) error {
	h.L.Error("Request Error",
		zap.String("method", req.Method), zap.String("path", req.Filepath),
		zap.Error(err))
	if os.IsNotExist(err) {
		return sftp.ErrSshFxNoSuchFile
	} else if os.IsPermission(err) {
		return sftp.ErrSshFxPermissionDenied
	}
	return sftp.ErrSSHFxFailure
}

func (h fileSystemHandlers) Fileread(req *sftp.Request) (io.ReaderAt, error) {
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

func (h fileSystemHandlers) Filewrite(req *sftp.Request) (io.WriterAt, error) {
	switch req.Method {
	case "Put", "Open":
		return nil, sftp.ErrSshFxPermissionDenied
	default:
		return nil, sftp.ErrSshFxOpUnsupported
	}
}

func (h fileSystemHandlers) Filecmd(req *sftp.Request) error {
	switch req.Method {
	case "Setstat", "Rename", "Rmdir", "Mkdir", "Link", "Symlink", "Remove":
		return sftp.ErrSshFxPermissionDenied
	default:
		return sftp.ErrSshFxOpUnsupported
	}
}

func (h fileSystemHandlers) Filelist(req *sftp.Request) (sftp.ListerAt, error) {
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
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}
