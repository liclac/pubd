package cliutil

import (
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/spf13/pflag"

	"github.com/liclac/pubd"
)

// Standard flags for constructing an http.FileSystem.
type FileSystemConfig struct {
	Path    string   `toml:"path"` // Normally given as os.Args[1].
	Exclude []string `toml:"exclude"`
}

// Defaults for FileSystemConfig.
func FileSystemDefaults() FileSystemConfig {
	pwd, _ := os.Getwd()
	return FileSystemConfig{Path: pwd}
}

func (c *FileSystemConfig) Flags(f *pflag.FlagSet) {
	f.StringSliceVarP(&c.Exclude, "exclude", "x", c.Exclude, "filenames/.gitignore patterns to exclude")
}

func (c FileSystemConfig) Build(fs billy.Filesystem) (http.FileSystem, error) {
	fs, err := fs.Chroot(c.Path)
	if err != nil {
		return nil, err
	}
	httpFS := pubd.FileSystem(fs)
	return pubd.FileSystemExclude(httpFS, c.Exclude), nil
}
