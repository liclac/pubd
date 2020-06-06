// +build !linux

package pubd

import (
	"fmt"
	"net"
	"runtime"
)

func ListenSystemd(network, addr string) ([]net.Listener, error) {
	return nil, fmt.Errorf("systemd support not built for %s", runtime.GOOS)
}
