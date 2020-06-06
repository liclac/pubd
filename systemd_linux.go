package pubd

import (
	"fmt"
	"net"

	"github.com/coreos/go-systemd/activation"
)

// Listen on file descriptors passed by systemd socket activation.
//
// Returns an error if no file descriptors were passed, which means either we're not
// running in a systemd unit, or it was started directly without the accompanying socket.
// The easiest way to avoid the latter is to add an explicit "Requires=" to the service;
// see the sample units provided in ./etc/systemd-*.
//
// network must be "systemd". addr is reserved for future arguments and must be empty.
//
// See also: https://www.freedesktop.org/software/systemd/man/systemd.socket.html
func ListenSystemd(network, addr string) ([]net.Listener, error) {
	if network != "systemd" {
		return nil, fmt.Errorf("expected network 'systemd', not '%s'", network)
	}
	if len(addr) > 0 {
		return nil, fmt.Errorf("no arguments defined, but got '%s'; if you have an idea for an argument, we'd love to hear it: https://github.com/liclac/pubd/issues", addr)
	}

	// We use Files() rather than Listeners(), because we don't want to swallow errors.
	files := activation.Files(true)
	listeners := make([]net.Listener, len(files))
	for i, f := range files {
		l, err := net.FileListener(f)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %s: %w", i, f.Name(), err)
		}
		listeners[i] = l
	}

	if len(listeners) == 0 {
		return nil, fmt.Errorf("not socket activated; if running in a systemd service, maybe you need to add a 'Requires=foo.socket' dependency?")
	}
	return listeners, nil
}
