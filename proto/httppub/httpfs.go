package httppub

import (
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5"
)

// Thin adapter between billy.Filesystem and http.FileSystem.
type httpFileSystem struct{ billy.Filesystem }

// Wraps a billy.Filesystem in a http.FileSystem. The wrapped filesystem should be "chrooted",
// in the sense that fs.Readdir("/") should read the root of the tree to be served.
func FileSystem(fs billy.Filesystem) http.FileSystem {
	return httpFileSystem{fs}
}

func (fs httpFileSystem) Open(name string) (http.File, error) {
	// Note: It's incorrect to Open() a directory with billy, and it fails with memfs.
	// Instead, we Stat() the path and return either a httpDir or a httpFile.
	info, err := fs.Filesystem.Stat(name)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return httpDir{name, info, fs.Filesystem}, nil
	}
	f, err := fs.Filesystem.Open(name)
	return httpFile{f, name, info, fs.Filesystem}, err
}

// Wraps a billy.Filesystem and a directory path in a http.File.
type httpDir struct {
	name string
	info os.FileInfo
	fs   billy.Filesystem
}

func (d httpDir) Stat() (os.FileInfo, error) { return d.info, nil }
func (d httpDir) Close() error               { return nil }

// None of these are valid operations on a directory.
func (d httpDir) Seek(int64, int) (int64, error) { return 0, os.ErrInvalid }
func (d httpDir) Read([]byte) (int, error)       { return 0, os.ErrInvalid }

// Note: os.File's Readdir can paginate, but Billy's just returns everything in one go.
// If this causes issues, adding pagination support is trivial; f.readdir = f.readdir[:num] etc.
func (d httpDir) Readdir(num int) ([]os.FileInfo, error) {
	return d.fs.ReadDir(d.name)
}

// Wraps a billy.File in a http.File.
type httpFile struct {
	billy.File
	name string
	info os.FileInfo
	fs   billy.Filesystem
}

func (f httpFile) Stat() (os.FileInfo, error) { return f.info, nil }

// Can't Readdir something that isn't a directory.
func (f httpFile) Readdir(num int) ([]os.FileInfo, error) { return nil, os.ErrInvalid }
