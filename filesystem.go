package pubd

import (
	"net/http"
	"os"

	"github.com/gobwas/glob"
	"github.com/spf13/pflag"
)

type FileSystemConfig struct {
	Path    string   `toml:"path"` // Normally given as os.Args[1].
	Exclude []string `toml:"exclude"`
}

func (c *FileSystemConfig) Flags(f *pflag.FlagSet) {
	f.StringSliceVarP(&c.Exclude, "exclude", "x", c.Exclude, "filenames or globs to exclude")
}

// Returns a filtered http.FileSystem at the given path.
func (c FileSystemConfig) Build() (http.FileSystem, error) {
	filter, err := c.filter()
	return fileSystem{filter, http.Dir(c.Path)}, err
}

func (c FileSystemConfig) filter() (fileSystemFilter, error) {
	filter := fileSystemFilter{Exclude: make([]glob.Glob, len(c.Exclude))}
	for i, expr := range c.Exclude {
		g, err := glob.Compile("**" + expr + "**")
		if err != nil {
			return filter, err
		}
		filter.Exclude[i] = g
	}
	return filter, nil
}

type fileSystemFilter struct {
	Exclude []glob.Glob
}

func (c fileSystemFilter) IsAllowed(name string) bool {
	for _, g := range c.Exclude {
		if g.Match(name) {
			return false
		}
	}
	return true
}

// To filter an http.FileSystem, we need to build a thin wrapper around it.
// It returns 404 for any excluded files, and returns a thin wrapper around
// opened directories that filters Readdir().
type fileSystem struct {
	fileSystemFilter
	fs http.FileSystem
}

func (fs fileSystem) Open(name string) (http.File, error) {
	if !fs.IsAllowed(name) {
		return nil, os.ErrNotExist
	}
	f, err := fs.fs.Open(name)
	if err != nil {
		return f, err
	}
	// Note: It's tempting to just skip the stat() call and always return
	// a wrapper, but that would disable sendfile() optimisations on unix
	// platforms, which is a far greater loss than any number of stat()s.
	if info, err := f.Stat(); err != nil {
		return f, err
	} else if info.IsDir() {
		return fileSystemDir{fs.fileSystemFilter, f}, nil
	}
	return f, nil
}

// As mentioned in the comment on fileSystem, fileSystemDir filters excluded
// files from directory listings by omitting them from Readdir().
type fileSystemDir struct {
	fileSystemFilter
	http.File
}

func (d fileSystemDir) Readdir(num int) ([]os.FileInfo, error) {
	infos, err := d.File.Readdir(num) // Careful - don't recurse!
	if err != nil {
		return infos, err
	}
	out := make([]os.FileInfo, 0, len(infos))
	for _, info := range infos {
		if d.IsAllowed(info.Name()) {
			out = append(out, info)
		}
	}
	// A literal reading of the docs suggests that it's semantically
	// incorrect to return an empty slice without an accompanying error.
	if len(out) == 0 {
		return d.Readdir(num) // Deliberately recursive!
	}
	return out, nil
}
