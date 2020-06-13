package sshpub

import (
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
