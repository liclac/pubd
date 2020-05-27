package cmd

import (
	"net/http"

	"github.com/liclac/pubd"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/pflag"
)

// Standard flags for constructing an http.FileSystem.
type FileSystemConfig struct {
	Path    string   `toml:"path"` // Normally given as os.Args[1].
	Exclude []string `toml:"exclude"`
}

func (c *FileSystemConfig) Flags(f *pflag.FlagSet) {
	f.StringSliceVarP(&c.Exclude, "exclude", "x", c.Exclude, "filenames/.gitignore patterns to exclude")
}

func (c FileSystemConfig) Build() http.FileSystem {
	fs := pubd.FileSystem(osfs.New(c.Path))
	return pubd.FileSystemExclude(fs, c.Exclude)
}
