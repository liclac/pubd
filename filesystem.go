package pubd

import (
	"os"
	"sort"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Sorts a slice of FileInfo structs in alphabetical order.
func SortFileInfos(infos []os.FileInfo) {
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name() < infos[j].Name() })
}

type filteredFileSystem struct {
	billy.Filesystem
	filter func(string, os.FileInfo) bool
}

// Returns a filesystem which includes only files for which `filter` returns true.
func FileSystemFilter(fs billy.Filesystem, filter func(string, os.FileInfo) bool) billy.Filesystem {
	return filteredFileSystem{fs, filter}
}

// Retuns a filesystem which excludes files matching the given .gitignore expressions.
func FileSystemExclude(fs billy.Filesystem, exprs []string) billy.Filesystem {
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

	return FileSystemFilter(fs, func(path string, info os.FileInfo) bool {
		return !ignored.Match(strings.Split(strings.Trim(path, "/"), "/"), info.IsDir())
	})
}
