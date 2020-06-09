package main

import (
	"path"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liclac/pubd/cliutil"
)

func mkTestFS(t *testing.T, files map[string]string) billy.Filesystem {
	fs := memfs.New()
	require.NoError(t, fs.MkdirAll("/", 0000))
	for name, contents := range files {
		require.NoError(t, fs.MkdirAll(path.Dir(name), 0000))
		require.NoError(t, util.WriteFile(fs, name, []byte(contents), 0000))
	}
	return fs
}

func TestParse(t *testing.T) {
	type FSC = cliutil.FileSystemConfig // These lines are getting too long.

	testdata := map[string]Config{
		"0":                              {},
		"0 www":                          {FileSystemConfig: FSC{Path: "www"}},
		"0 -a localhost:9999":            {Addr: "localhost:9999"},
		"0 --addr=localhost:9999":        {Addr: "localhost:9999"},
		"0 -P ~liclac":                   {Prefix: "~liclac"},
		"0 --prefix=~liclac":             {Prefix: "~liclac"},
		"0 -x .git":                      {FileSystemConfig: FSC{Exclude: []string{".git"}}},
		"0 --exclude=.git":               {FileSystemConfig: FSC{Exclude: []string{".git"}}},
		"0 -x .git -x tmp":               {FileSystemConfig: FSC{Exclude: []string{".git", "tmp"}}},
		"0 --exclude=.git --exclude=tmp": {FileSystemConfig: FSC{Exclude: []string{".git", "tmp"}}},
	}
	for in, out := range testdata {
		if out.Addr == "" {
			out.Addr = "localhost:8888"
		}
		if out.FileSystemConfig.Path == "" {
			out.FileSystemConfig.Path = cliutil.FileSystemDefaults().Path
		}
		t.Run(in, func(t *testing.T) {
			cfg, err := Parse(memfs.New(), strings.Split(in, " "))
			require.NoError(t, err)
			assert.Equal(t, out, cfg)
		})
	}
}
