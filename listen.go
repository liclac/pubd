package pubd

import (
	"net"
	"strings"
)

// Opens a listener, rawAddr can be an address (localhost, 127.0.0.1:1337),
// or a network/address pair (tcp/localhost:1337, unix//tmp/pubd.sock).
func Listen(rawAddr string) (net.Listener, error) {
	network, addr := SplitAddr(rawAddr)
	return net.Listen(network, addr)
}

// Splits the argument to Listen into a network/address pair for net.Listen.
func SplitAddr(rawAddr string) (network, addr string) {
	if idx := strings.Index(rawAddr, "/"); idx > -1 {
		return rawAddr[:idx], rawAddr[idx+1:]
	}
	return "tcp", rawAddr
}
