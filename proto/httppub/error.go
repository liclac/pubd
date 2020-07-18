package httppub

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

var ErrMethodNotAllowed = errors.New("method not allowed")

// Guesses an appropriate status code for an error.
func ErrorCode(err error) int {
	if os.IsNotExist(err) {
		return http.StatusNotFound
	} else if errors.Is(err, ErrMethodNotAllowed) {
		return http.StatusMethodNotAllowed
	}
	return http.StatusInternalServerError
}

// Responds with an error page, in the form "404 Not Found".
func RenderError(rw http.ResponseWriter, req *http.Request, err error) {
	status := ErrorCode(err)
	rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
	rw.WriteHeader(status)
	fmt.Fprintln(rw, status, http.StatusText(status))
}
