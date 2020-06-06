package pubd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO: Test this by emulating the environment variables used by sd_listen_fds().
// https://www.freedesktop.org/software/systemd/man/sd_listen_fds.html
func TestListen_Systemd(t *testing.T) {
	t.Run("Extra Argument", func(t *testing.T) {
		_, err := Listen("systemd/blah")
		assert.EqualError(t, err, "listen/systemd: no arguments defined, but got 'blah'; if you have an idea for an argument, we'd love to hear it: https://github.com/liclac/pubd/issues")
	})

	t.Run("Not Socket Activated", func(t *testing.T) {
		_, err := Listen("systemd/")
		assert.EqualError(t, err, "listen/systemd: not socket activated; if running in a systemd service, maybe you need to add a 'Requires=foo.socket' dependency?")
	})
}
