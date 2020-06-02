package httppub

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode(t *testing.T) {
	testdata := map[string]struct {
		Err  error
		Code int
	}{
		"fmt.Errrof":          {fmt.Errorf("im gay"), http.StatusInternalServerError},
		"os.ErrNotExist":      {os.ErrNotExist, http.StatusNotFound},
		"ErrMethodNotAllowed": {ErrMethodNotAllowed, http.StatusMethodNotAllowed},
	}
	for name, tdata := range testdata {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tdata.Code, ErrorCode(tdata.Err))
		})
	}
}
