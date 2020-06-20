package sshpub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeUint32(t *testing.T) {
	testdata := map[string]struct {
		OK   bool
		Data []byte
		V    uint32
		Rest []byte
	}{
		"len=0": {false, []byte{}, 0, []byte{}},
		"len=1": {false, []byte{0x11}, 0, []byte{0x11}},
		"len=2": {false, []byte{0x11, 0x22}, 0, []byte{0x11, 0x22}},
		"len=3": {false, []byte{0x11, 0x22, 0x33}, 0, []byte{0x11, 0x22, 0x33}},
		"len=4": {true, []byte{0x11, 0x22, 0x33, 0x44}, 0x11223344, []byte{}},
		"len=5": {true, []byte{0x11, 0x22, 0x33, 0x44, 0x55}, 0x11223344, []byte{0x55}},

		"\x00\x00\x00\x01": {true, []byte{0x00, 0x00, 0x00, 0x01}, 1, []byte{}},
		"\x00\x00\x00\x02": {true, []byte{0x00, 0x00, 0x00, 0x02}, 2, []byte{}},
	}
	for name, tdata := range testdata {
		t.Run(name, func(t *testing.T) {
			v, rest, ok := DecodeUint32(tdata.Data)
			assert.Equal(t, tdata.V, v)
			assert.Equal(t, tdata.Rest, rest)
			assert.Equal(t, tdata.OK, ok)
		})
	}
}

func TestDecodeString(t *testing.T) {
	testdata := map[string]struct {
		OK   bool
		S    string
		Rest []byte
	}{
		"":                       {false, "", []byte{}},
		"\x00":                   {false, "", []byte{0x00}},
		"\x00\x00":               {false, "", []byte{0x00, 0x00}},
		"\x00\x00\x00":           {false, "", []byte{0x00, 0x00, 0x00}},
		"\x00\x00\x00\x00":       {true, "", []byte{}},
		"\x00\x00\x00\x01":       {false, "", []byte{}},
		"\x00\x00\x00\x01s":      {true, "s", []byte{}},
		"\x00\x00\x00\x01sftp":   {true, "s", []byte{'f', 't', 'p'}},
		"\x00\x00\x00\x02sftp":   {true, "sf", []byte{'t', 'p'}},
		"\x00\x00\x00\x03sftp":   {true, "sft", []byte{'p'}},
		"\x00\x00\x00\x04sftp":   {true, "sftp", []byte{}},
		"\x00\x00\x00\x04sftpab": {true, "sftp", []byte{'a', 'b'}},
		"\xFF\xFF\xFF\xFF":       {false, "", []byte{}},
	}
	for in, out := range testdata {
		t.Run(fmt.Sprintf("%q", in), func(t *testing.T) {
			s, rest, ok := DecodeString([]byte(in))
			assert.Equal(t, out.S, s)
			assert.Equal(t, out.Rest, rest)
			assert.Equal(t, out.OK, ok)
		})
	}
}
