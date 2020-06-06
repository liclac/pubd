package pubd

import (
	"fmt"
	"net"
	"strings"
)

// Splits the argument to Listen into a network/address pair for net.Listen.
func SplitAddr(rawAddr string) (network, addr string) {
	if idx := strings.Index(rawAddr, "/"); idx > -1 {
		return rawAddr[:idx], rawAddr[idx+1:]
	}
	return "tcp", rawAddr
}

// Listens on an address. rawAddr can be an address (localhost, 127.0.0.1:1337), or a
// network/address pair (tcp/localhost:1337, unix//tmp/pubd.sock).
//
// Special network types:
// - systemd/:
//   Use systemd socket activation. Returns an error if not running from a socket unit.
//   The best way to avoid this is to add an explicit Requires= to the unit definition.
func Listen(rawAddr string) ([]net.Listener, error) {
	network, addr := SplitAddr(rawAddr)
	switch network {
	case "systemd": // systemd socket activation.
		ls, err := ListenSystemd(network, addr)
		if err != nil {
			return nil, fmt.Errorf("listen/%s: %w", network, err)
		}
		return ls, nil
	default: // Regular ol' net.Listen().
		l, err := net.Listen(network, addr)
		if err != nil {
			return nil, fmt.Errorf("listen/%s: %w", network, err)
		}
		return []net.Listener{l}, nil
	}
}
