package httppub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNegotiate(t *testing.T) {
	testdata := map[string]string{
		"":                            "text/plain",
		"text/plain":                  "text/plain",
		"text/html":                   "text/html",
		"text/html, text/plain":       "text/html",
		"text/html;q=0.9, text/plain": "text/plain",
		"text/plain, text/html;q=0.9": "text/plain",
	}
	for accept, contentType := range testdata {
		t.Run(`"`+accept+`"`, func(t *testing.T) {
			assert.Equal(t, contentType, Negotiate(accept, "text/plain", "text/html"))
		})
	}
}

func Test_parseAcceptValue(t *testing.T) {
	testdata := map[string]struct {
		Type   string
		Weight float64
	}{
		"":                                     {"", 1.0},
		"text/plain":                           {"text/plain", 1.0},
		"text/plain;q=0.9":                     {"text/plain", 0.9},
		"text/plain; q=0.9":                    {"text/plain", 0.9},
		"text/plain; q=0.9; charset=utf-8":     {"text/plain", 0.9},
		"text/plain; q = 0.9; charset = utf-8": {"text/plain", 0.9},
	}
	for input, expected := range testdata {
		t.Run(`"`+input+`"`, func(t *testing.T) {
			contentType, weight := parseAcceptValue([]byte(input))
			assert.Equal(t, expected.Type, string(contentType))
			assert.Equal(t, expected.Weight, weight)
		})
	}
}
