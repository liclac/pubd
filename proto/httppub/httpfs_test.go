package httppub

import (
	"io/ioutil"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem(t *testing.T) {
	t.Run(".git", func(t *testing.T) {
		baseFS := memfs.New()
		require.NoError(t, baseFS.MkdirAll("/.git", 0000))
		require.NoError(t, util.WriteFile(baseFS, "/.git/HEAD", []byte("ref: refs/heads/master"), 0000))
		fs := FileSystem(baseFS)

		d, err := fs.Open("/.git")
		require.NoError(t, err)
		assert.IsType(t, httpDir{}, d)

		t.Run("Stat", func(t *testing.T) {
			info, err := d.Stat()
			require.NoError(t, err)
			assert.Equal(t, ".git", info.Name())
			assert.True(t, info.IsDir())
		})

		t.Run("Readdir", func(t *testing.T) {
			infos, err := d.Readdir(-1)
			require.NoError(t, err)
			if assert.Len(t, infos, 1) {
				assert.Equal(t, "HEAD", infos[0].Name())
				assert.Equal(t, int64(22), infos[0].Size())
			}
		})

		t.Run("Read", func(t *testing.T) {
			_, err := ioutil.ReadAll(d)
			assert.EqualError(t, err, "invalid argument")
		})

		t.Run("Close", func(t *testing.T) {
			assert.NoError(t, d.Close())
		})
	})

	t.Run("HEAD", func(t *testing.T) {
		baseFS := memfs.New()
		require.NoError(t, baseFS.MkdirAll("/.git", 0000))
		require.NoError(t, util.WriteFile(baseFS, "/.git/HEAD", []byte("ref: refs/heads/master"), 0000))
		fs := FileSystem(baseFS)

		f, err := fs.Open("/.git/HEAD")
		require.NoError(t, err)
		assert.IsType(t, httpFile{}, f)

		t.Run("Stat", func(t *testing.T) {
			info, err := f.Stat()
			require.NoError(t, err)
			assert.Equal(t, "HEAD", info.Name())
			assert.Equal(t, int64(22), info.Size())
			assert.False(t, info.IsDir())
		})

		t.Run("Readdir", func(t *testing.T) {
			_, err := f.Readdir(-1)
			assert.EqualError(t, err, "invalid argument")
		})

		t.Run("Read", func(t *testing.T) {
			data, err := ioutil.ReadAll(f)
			assert.NoError(t, err)
			assert.Equal(t, "ref: refs/heads/master", string(data))
		})

		assert.NoError(t, f.Close())
	})

	t.Run("Nonexistent", func(t *testing.T) {
		fs := FileSystem(memfs.New())
		_, err := fs.Open("/nonexistent")
		assert.EqualError(t, err, "file does not exist")
	})
}
