package httppub

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liclac/pubd/testutil"
)

func TestSimpleIndexEmpty(t *testing.T) {
	idx := SimpleIndex(IndexConfig{})
	fs := memfs.New()

	t.Run("Plain", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, "", rw.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "text/html")
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, "<pre>\n</pre>", rw.Body.String())
	})
}

func TestSimpleIndex(t *testing.T) {
	idx := SimpleIndex(IndexConfig{})
	fs := memfs.New()

	t.Run("Plain", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{
			testutil.FileInfo{FName: ".git", FIsDir: true},
			testutil.FileInfo{FName: "index.html"},
		}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, ".git/\nindex.html\n", rw.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "text/html")
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{
			testutil.FileInfo{FName: ".git", FIsDir: true},
			testutil.FileInfo{FName: "index.html"},
		}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, `<pre>
<a href=".git/">.git/</a>
<a href="index.html">index.html</a>
</pre>`, rw.Body.String())
	})
}

func TestSimpleIndexREADME(t *testing.T) {
	idx := SimpleIndex(IndexConfig{READMEs: []string{"README.txt"}})
	fs := memfs.New()
	require.NoError(t, util.WriteFile(fs, "/README.txt", []byte("this is a readme"), 0666))

	t.Run("Plain", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{
			testutil.FileInfo{FName: ".git", FIsDir: true},
			testutil.FileInfo{FName: "index.html"},
			testutil.FileInfo{FName: "README.txt"},
		}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, ".git/\nindex.html\nREADME.txt\n\nthis is a readme", rw.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "text/html")
		assert.NoError(t, idx.Render(rw, req, fs, []os.FileInfo{
			testutil.FileInfo{FName: ".git", FIsDir: true},
			testutil.FileInfo{FName: "index.html"},
			testutil.FileInfo{FName: "README.txt"},
		}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, `<pre>
<a href=".git/">.git/</a>
<a href="index.html">index.html</a>
<a href="README.txt">README.txt</a>

this is a readme</pre>`, rw.Body.String())
	})
}
