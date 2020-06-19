package pubd

import (
	"io/ioutil"
	"os"
	"testing"

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
		"index.html": {
			"/index.txt":         true,
			"/index.html":        false,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": true,
		},
		"*.html": {
			"/index.txt":         true,
			"/index.html":        false,
			"/about.html":        false,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
		"subdir": {
			"/index.txt":         true,
			"/index.html":        true,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
		"subdir/*.html": {
			"/index.txt":         true,
			"/index.html":        true,
			"/about.html":        true,
			"/subdir/index.html": false,
			"/subdir/about.html": false,
		},
	}
	for pattern, pathdata := range testdata {
		t.Run(`"`+pattern+`"`, func(t *testing.T) {
			for path, allowed := range pathdata {
				t.Run(`"`+path+`"`, func(t *testing.T) {
					baseFS := memfs.New()
					require.NoError(t, baseFS.MkdirAll("subdir", 0000))
					for path := range pathdata {
						require.NoError(t, util.WriteFile(baseFS, path, []byte(path), 0000))
					}

					fs := FileSystemExclude(baseFS, []string{pattern})
					f, err := fs.Open(path)
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
}
