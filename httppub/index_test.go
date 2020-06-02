package httppub

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/liclac/pubd/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSimpleIndexEmpty(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, SimpleIndex()(rw, req, testutil.FileInfo{FName: "dir"}, []os.FileInfo{}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, "", rw.Body.String())
	})

	t.Run("HTML", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "text/html")
		assert.NoError(t, SimpleIndex()(rw, req, testutil.FileInfo{FName: "dir"}, []os.FileInfo{}))
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/html; charset=utf-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, "<pre>\n</pre>", rw.Body.String())
	})
}

func TestSimpleIndex(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		assert.NoError(t, SimpleIndex()(rw, req, testutil.FileInfo{FName: "dir"}, []os.FileInfo{
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
		assert.NoError(t, SimpleIndex()(rw, req, testutil.FileInfo{FName: "dir"}, []os.FileInfo{
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
