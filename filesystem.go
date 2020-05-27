package pubd

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/plumbing/format/gitignore"
	"github.com/shurcooL/httpfs/filter"
)

// Retuns a filesystem which excludes files matching the given .gitignore expressions.
func FileSystemExclude(fs http.FileSystem, exprs []string) http.FileSystem {
	if len(exprs) == 0 {
		return fs
	}

	// The second argument here is the "domain" of the pattern, eg. the root directory under which
	// it applies, as a slice of path segments ("/srv/www" -> ["srv", "www"]). In this case, we're
	// working with a "chrooted" vfs, and only one set of patterns, so it's always "/" -> [] = nil.
	patterns := make([]gitignore.Pattern, len(exprs))
	for i, expr := range exprs {
		patterns[i] = gitignore.ParsePattern(expr, nil)
	}
	ignored := gitignore.NewMatcher(patterns)

	return filter.Keep(fs, func(path string, info os.FileInfo) bool {
		return !ignored.Match(strings.Split(strings.Trim(path, "/"), "/"), info.IsDir())
	})
}

// Thin adapter between billy.Filesystem and http.FileSystem.
type httpFileSystem struct{ billy.Filesystem }

// Wraps a billy.Filesystem in a http.FileSystem. The wrapped filesystem should be "chrooted",
// in the sense that fs.Readdir("/") should read the root of the tree to be served.
func FileSystem(fs billy.Filesystem) http.FileSystem {
	return httpFileSystem{fs}
}

func (fs httpFileSystem) Open(name string) (http.File, error) {
	f, err := fs.Filesystem.Open(name)
	return httpFile{f, fs.Filesystem, name}, err
}

// Wraps a billy.File in a http.File.
type httpFile struct {
	billy.File
	fs   billy.Filesystem
	name string
}

// Note: os.File's Readdir can paginate, but Billy's just returns everything in one go.
// If this causes issues, adding pagination support is trivial; f.readdir = f.readdir[:num] etc.
func (f httpFile) Readdir(num int) ([]os.FileInfo, error) {
	return f.fs.ReadDir(f.name)
}

// This is on os.File, but on billy.Filesystem.
func (f httpFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(f.name)
}
