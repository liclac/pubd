package pubd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitAddr(t *testing.T) {
	testdata := map[string][2]string{
		"localhost":           {"tcp", "localhost"},
		"localhost:1337":      {"tcp", "localhost:1337"},
		"tcp/localhost":       {"tcp", "localhost"},
		"tcp/localhost:1337":  {"tcp", "localhost:1337"},
		"unix//tmp/pubd.sock": {"unix", "/tmp/pubd.sock"},
		"/localhost":          {"", "localhost"},
		"/":                   {"", ""},
		"":                    {"tcp", ""},
	}
	for rawAddr, out := range testdata {
		t.Run(rawAddr, func(t *testing.T) {
			net, addr := SplitAddr(rawAddr)
			assert.Equal(t, out[0], net)
			assert.Equal(t, out[1], addr)
		})
	}
}
