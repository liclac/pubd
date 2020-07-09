package httppub

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerNoIndex(t *testing.T) {
	fs := memfs.New()
	require.NoError(t, fs.MkdirAll("/.git", 0000))
	require.NoError(t, util.WriteFile(fs, "/.git/HEAD", []byte("ref: refs/heads/master"), 0000))

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

func TestHandlerSimpleIndex(t *testing.T) {
	fs := memfs.New()
	require.NoError(t, fs.MkdirAll("/.git", 0000))
	require.NoError(t, util.WriteFile(fs, "/.git/HEAD", []byte("ref: refs/heads/master"), 0000))

	srv := httptest.NewServer(Handler(fs, SimpleIndex(), nil))
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
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", rsp.Header.Get("Content-Type"))
		body, err := ioutil.ReadAll(rsp.Body)
		require.NoError(t, err)
		assert.Equal(t, ".git/\n", string(body))
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

func Test_cleanPath(t *testing.T) {
	testdata := map[string]string{
		"":            "/",
		"/":           "/",
		"index.html":  "/index.html",
		"/index.html": "/index.html",
		"/dir":        "/dir",
		"/dir/":       "/dir/",
		"//dir":       "/dir",
		"//dir//":     "/dir/",
		"/dir/a":      "/dir/a",
		"/dir/a/..":   "/dir",
		"/dir/a/../":  "/dir/",
	}
	for in, out := range testdata {
		assert.Equal(t, out, cleanPath(in))
	}
}

func Test_redirectForCanon(t *testing.T) {
	testdata := []struct {
		path  string
		isDir bool
		out   string
	}{
		{"/", true, ""},
		{"/", false, ".."},
		{"/.git", true, ".git/"},
		{"/.git", false, ""},
		{"/.git/", true, ""},
		{"/.git/", false, "../.git"},
		{"/.git/HEAD", false, ""},
		{"/.git/HEAD", true, "HEAD/"},
		{"/.git/HEAD/", false, "../HEAD"},
		{"/.git/HEAD/", true, ""},
	}
	for _, tdata := range testdata {
		t.Run(fmt.Sprintf("%s (%v)", tdata.path, tdata.isDir), func(t *testing.T) {
			assert.Equal(t, tdata.out, redirectForCanon(tdata.path, tdata.isDir))
		})
	}
}
