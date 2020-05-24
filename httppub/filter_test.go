package httppub

import (
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
