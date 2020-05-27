package main

import (
	"path"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/liclac/pubd/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mkTestFS(t *testing.T, files map[string]string) billy.Filesystem {
	fs := memfs.New()
	for name, contents := range files {
		require.NoError(t, fs.MkdirAll(path.Dir(name), 0000))
		require.NoError(t, util.WriteFile(fs, name, []byte(contents), 0000))
	}
	return fs
}

func TestParse(t *testing.T) {
	type FSC = cmd.FileSystemConfig // These lines are getting too long.

	testdata := map[string]Config{
		"":                             {},
		"www":                          {FileSystemConfig: FSC{Path: "www"}},
		"-a localhost:9999":            {Addr: "localhost:9999"},
		"--addr=localhost:9999":        {Addr: "localhost:9999"},
		"-q":                           {Quiet: true},
		"--quiet":                      {Quiet: true},
		"-P ~liclac":                   {Prefix: "~liclac"},
		"--prefix=~liclac":             {Prefix: "~liclac"},
		"-x .git":                      {FileSystemConfig: FSC{Exclude: []string{".git"}}},
		"--exclude=.git":               {FileSystemConfig: FSC{Exclude: []string{".git"}}},
		"-x .git -x tmp":               {FileSystemConfig: FSC{Exclude: []string{".git", "tmp"}}},
		"--exclude=.git --exclude=tmp": {FileSystemConfig: FSC{Exclude: []string{".git", "tmp"}}},
	}
	for in, out := range testdata {
		if out.Addr == "" {
			out.Addr = "localhost:8080"
		}
		t.Run(in, func(t *testing.T) {
			cfg, err := Parse(strings.Split(in, " "))
			require.NoError(t, err)
			assert.Equal(t, out, cfg)
		})
	}
}
