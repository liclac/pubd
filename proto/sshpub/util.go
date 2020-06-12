package sshpub

import (
	"io/ioutil"

	"github.com/go-git/go-billy/v5"
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
