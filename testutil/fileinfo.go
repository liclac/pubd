package testutil

import (
	"os"
	"time"
)

var _ os.FileInfo = FileInfo{}

type FileInfo struct {
	FName    string
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
	FIsDir   bool
	FSys     interface{}
}

func (i FileInfo) Name() string       { return i.FName }
func (i FileInfo) Size() int64        { return i.FSize }
func (i FileInfo) Mode() os.FileMode  { return i.FMode }
func (i FileInfo) ModTime() time.Time { return i.FModTime }
func (i FileInfo) IsDir() bool        { return i.FIsDir }
func (i FileInfo) Sys() interface{}   { return i.FSys }
