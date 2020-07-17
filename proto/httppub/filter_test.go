package httppub

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanPrefix(t *testing.T) {
	testdata := map[string]string{
		"":          "",
		"/":         "",
		"//":        "",
		"///":       "",
		"////":      "",
		"prefix":    "/prefix",
		"prefix/":   "/prefix",
		"/prefix/":  "/prefix",
		"/prefix":   "/prefix",
		"pre/fix":   "/pre/fix",
		"pre/fix/":  "/pre/fix",
		"/pre/fix/": "/pre/fix",
		"/pre/fix":  "/pre/fix",
	}
	for input, output := range testdata {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, output, CleanPrefix(input))
		})
	}
}

func TestWithPrefix(t *testing.T) {
	testdata := map[string]struct {
		Status   int
		Location string
	}{
		"/":          {http.StatusNotFound, ""},
		"/prefix":    {http.StatusMovedPermanently, "prefix/"},
		"/prefix/":   {http.StatusOK, ""},
		"/prefix/hi": {http.StatusOK, ""},
	}
	handler := WithPrefix("/prefix",
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
	for path, tdata := range testdata {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)
			assert.Equal(t, tdata.Status, rw.Code)
			assert.Equal(t, tdata.Location, rw.Header().Get("Location"))
		})
	}
}
