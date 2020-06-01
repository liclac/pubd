package pubd

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ os.FileInfo = testFileInfo{}

type testFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func (i testFileInfo) Name() string       { return i.name }
func (i testFileInfo) Size() int64        { return i.size }
func (i testFileInfo) Mode() os.FileMode  { return i.mode }
func (i testFileInfo) ModTime() time.Time { return i.modTime }
func (i testFileInfo) IsDir() bool        { return i.isDir }
func (i testFileInfo) Sys() interface{}   { return i.sys }

func TestSortFileInfos(t *testing.T) {
	infos := []os.FileInfo{
		testFileInfo{name: "a"},
		testFileInfo{name: "c"},
		testFileInfo{name: "ab"},
		testFileInfo{name: "d"},
	}
	SortFileInfos(infos)
	assert.Equal(t, []os.FileInfo{
		testFileInfo{name: "a"},
		testFileInfo{name: "ab"},
		testFileInfo{name: "c"},
		testFileInfo{name: "d"},
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

					fs := FileSystemExclude(FileSystem(baseFS), []string{pattern})
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
