package httppub

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liclac/pubd"
)

func TestHandlerNoIndex(t *testing.T) {
	baseFS := memfs.New()
	require.NoError(t, baseFS.MkdirAll("/.git", 0000))
	require.NoError(t, util.WriteFile(baseFS, "/.git/HEAD", []byte("ref: refs/heads/master"), 0000))
	fs := pubd.FileSystem(baseFS)

	srv := httptest.NewServer(Handler(fs, nil, nil))
	defer srv.Close()

	t.Run("POST /", func(t *testing.T) {
		rsp, err := http.Post(srv.URL, "", nil)
		require.NoError(t, err)
		defer rsp.Body.Close()
		assert.Equal(t, http.StatusMethodNotAllowed, rsp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", rsp.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(rsp.Body)
		require.NoError(t, err)
		assert.Equal(t, "405 Method Not Allowed\n", string(body))
	})

	t.Run("GET /", func(t *testing.T) {
		rsp, err := http.Get(srv.URL)
		require.NoError(t, err)
		defer rsp.Body.Close()
		assert.Equal(t, http.StatusNotFound, rsp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", rsp.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(rsp.Body)
		require.NoError(t, err)
		assert.Equal(t, "404 Not Found\n", string(body))
	})

	t.Run("GET /.git/HEAD", func(t *testing.T) {
		rsp, err := http.Get(srv.URL + "/.git/HEAD")
		require.NoError(t, err)
		defer rsp.Body.Close()
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", rsp.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(rsp.Body)
		require.NoError(t, err)
		assert.Equal(t, "ref: refs/heads/master", string(body))
	})
}
