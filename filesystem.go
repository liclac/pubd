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
	filter func(string, bool) bool
}

// Returns a filesystem which includes only files for which `filter` returns true.
func FileSystemFilter(fs billy.Filesystem, filter func(string, bool) bool) billy.Filesystem {
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

	return FileSystemFilter(fs, func(path string, isDir bool) bool {
		return !ignored.Match(strings.Split(strings.Trim(path, "/"), "/"), isDir)
	})
}

// Silly optimisations: try to evaluate the filter only once, and avoid hitting the disk
// to stat() upfront if if wouldn't be allowed anyhow.
func (fs filteredFileSystem) isAllowed(filename string) (bool, error) {
	if fs.filter(filename, false) {
		info, err := fs.Filesystem.Stat(filename)
		return info != nil && (!info.IsDir() || fs.filter(filename, true)), err
	} else if fs.filter(filename, true) {
		info, err := fs.Filesystem.Stat(filename)
		return info != nil && info.IsDir(), err
	} else {
		return false, nil
	}
}

func (fs filteredFileSystem) Open(filename string) (billy.File, error) {
	isAllowed, err := fs.isAllowed(filename)
	if err != nil {
		return nil, err
	} else if !isAllowed {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.Open(filename)
}

func (fs filteredFileSystem) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	isAllowed, err := fs.isAllowed(filename)
	if err != nil {
		return nil, err
	} else if !isAllowed {
		return nil, os.ErrNotExist
	}
	return fs.Filesystem.OpenFile(filename, flag, perm)
}

func (fs filteredFileSystem) ReadDir(filename string) ([]os.FileInfo, error) {
	isAllowed, err := fs.isAllowed(filename)
	if err != nil {
		return nil, err
	} else if !isAllowed {
		return nil, os.ErrNotExist
	}
	realInfos, err := fs.Filesystem.ReadDir(filename)
	if err != nil {
		return nil, err
	}
	filteredInfos := make([]os.FileInfo, 0, len(realInfos))
	for _, info := range realInfos {
		if fs.filter(info.Name(), info.IsDir()) {
			filteredInfos = append(filteredInfos, info)
		}
	}
	return filteredInfos, nil
}
