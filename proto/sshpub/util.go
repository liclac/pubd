package sshpub

import (
	"encoding/binary"
	"io/ioutil"

	"github.com/go-git/go-billy/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func LoadPrivateKey(fs billy.Filesystem, path string) (ssh.Signer, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}

// Encodes an RFC4251 "uint32" value; 4 big endian bytes.
func EncodeUint32(v uint32) []byte {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], v)
	return data[:]
}

// Decodes an RFC4251 "uint32" value; 4 big endian bytes.
// Returns the value, whether there was enough data, and the remainder.
func DecodeUint32(data []byte) (uint32, []byte, bool) {
	if len(data) < 4 {
		return 0, data, false
	}
	return binary.BigEndian.Uint32(data), data[4:], true
}

// Decodes an RFC4251 "string" value; a byte sequence prefixed with a uint32 length field.
// Returns the value, whether there was enough data, and the remainder.
func DecodeString(data []byte) (string, []byte, bool) {
	l, data, ok := DecodeUint32(data)
	if !ok || uint32(len(data)) < l {
		return "", data, false
	}
	return string(data[:l]), data[l:], true
}

// If the request wants a reply, decline it, else discard it. Log a warning if declining failed.
func lazyReply(L *zap.Logger, req *ssh.Request, ok bool) {
	if !req.WantReply {
		return
	}
	if err := req.Reply(ok, nil); err != nil {
		msg := "Error declining request"
		if ok {
			msg = "Error discarding request"
		}
		L.Warn(msg, zap.String("type", req.Type), zap.Error(err))
	}
}
