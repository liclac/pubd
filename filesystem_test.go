package pubd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liclac/pubd/testutil"
)

func TestSortFileInfos(t *testing.T) {
	infos := []os.FileInfo{
		testutil.FileInfo{FName: "a"},
		testutil.FileInfo{FName: "c"},
		testutil.FileInfo{FName: "ab"},
		testutil.FileInfo{FName: "d"},
	}
	SortFileInfos(infos)
	assert.Equal(t, []os.FileInfo{
		testutil.FileInfo{FName: "a"},
		testutil.FileInfo{FName: "ab"},
		testutil.FileInfo{FName: "c"},
		testutil.FileInfo{FName: "d"},
	}, infos)
}

func TestFileSystemExclude(t *testing.T) {
	testdata := map[string]map[string]bool{
		".*": {
			".git/HEAD":          false,
			"/index.txt":         true,
			"/index.html":        true,
			"/about.html":        true,
			"/subdir/index.html": true,
			"/subdir/about.html": true,
		},
		"index.html": {
			".git/HEAD":          true,
			"/index.txt":         true,
			"/index.html":        false,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": true,
		},
		"*.html": {
			".git/HEAD":          true,
			"/index.txt":         true,
			"/index.html":        false,
			"/about.html":        false,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
		"subdir": {
			".git/HEAD":          true,
			"/index.txt":         true,
			"/index.html":        true,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
		"subdir/*.html": {
			".git/HEAD":          true,
			"/index.txt":         true,
			"/index.html":        true,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
	}
	openers := map[string]func(billy.Basic, string) (billy.File, error){
		"Open": func(fs billy.Basic, name string) (billy.File, error) {
			return fs.Open(name)
		},
		"OpenFile": func(fs billy.Basic, name string) (billy.File, error) {
			return fs.OpenFile(name, os.O_RDONLY, 0000)
		},
	}
	for pattern, pathdata := range testdata {
		t.Run(`"`+pattern+`"`, func(t *testing.T) {
			for path, allowed := range pathdata {
				t.Run(`"`+path+`"`, func(t *testing.T) {
					for name, open := range openers {
						t.Run(name, func(t *testing.T) {
							baseFS := memfs.New()
							require.NoError(t, baseFS.MkdirAll(".git", 0000))
							require.NoError(t, baseFS.MkdirAll("subdir", 0000))
							for path := range pathdata {
								require.NoError(t, util.WriteFile(baseFS, path, []byte(path), 0000))
							}

							fs := FileSystemExclude(baseFS, []string{pattern})
							f, err := open(fs, path)
							if !allowed {
								assert.True(t, os.IsNotExist(err), "should return ErrNotExist")
							} else {
								require.NoError(t, err)
								data, err := ioutil.ReadAll(f)
								require.NoError(t, err)
								assert.Equal(t, path, string(data))
							}
						})
					}
				})
			}
		})
	}
}
