package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/cliutil"
	"github.com/liclac/pubd/proto/sftppub"
	"github.com/liclac/pubd/proto/sshpub"
)

const Usage = `usage: pubd-ssh [-F] [path]`

type ServerConfig struct {
	SFTP struct {
		Enable bool `toml:"enable"`
	} `toml:"sftp"`
}

type Config struct {
	Addr        string `toml:"addr"`
	HostKeyFile string `toml:"host-key-file"` // Path to host private key.
	ServerConfig
	cliutil.FileSystemConfig
	cliutil.LogConfig
}

func Parse(fs billy.Filesystem, args []string) (Config, error) {
	cfg := Config{Addr: "localhost:2222", FileSystemConfig: cliutil.FileSystemDefaults()}
	return cfg, cliutil.Configure(&cfg, &cfg.Path, func(f *pflag.FlagSet) {
		f.StringVarP(&cfg.Addr, "addr", "a", cfg.Addr, "listen address")
		f.BoolVarP(&cfg.SFTP.Enable, "sftp.enable", "F", cfg.SFTP.Enable, "enable SFTP access")
		f.StringVarP(&cfg.HostKeyFile, "host-key-file", "K", cfg.HostKeyFile, "path to host private key file")
		cfg.FileSystemConfig.Flags(f)
		cfg.LogConfig.Flags(f)
	}, Usage, args)
}

func Server(L *zap.Logger, fs billy.Filesystem, hostKey ssh.Signer, cfg ServerConfig) pubd.Server {
	var subSFTP sshpub.Subsystem
	if cfg.SFTP.Enable {
		subSFTP = sftppub.New(fs)
	}

	// Subsystems should be explicitly set to nil when disabled; an unset subsystem logs a warning.
	srv := sshpub.New(L, hostKey)
	srv.Subsystems = map[string]sshpub.Subsystem{
		"sftp": subSFTP,
	}
	return pubd.ServerFunc(func(ctx context.Context, l net.Listener) error {
		L.Info("Running", zap.Stringer("addr", l.Addr()))
		return srv.Serve(ctx, l)
	})
}

func Main(hostFS billy.Filesystem, args []string) error {
	cfg, err := Parse(hostFS, args)
	if err != nil {
		return err
	}
	if !cfg.SFTP.Enable {
		return errors.New("no transports enabled; try -F/--sftp.enable")
	}

	if cfg.HostKeyFile == "" {
		return errors.New("-K/--host-key-file is required, and can be generated with: `ssh-keygen -t ed25519`")
	}
	hostKeyPath, err := filepath.Abs(cfg.HostKeyFile)
	if err != nil {
		return fmt.Errorf("-K/--host-key-file: couldn't absolutise path to '%s': %w", cfg.HostKeyFile, err)
	}
	hostKey, err := sshpub.LoadPrivateKey(hostFS, hostKeyPath)
	if err != nil {
		return err
	}

	fs, err := cfg.Build(hostFS) // TODO: This is a weird function name.
	if err != nil {
		return err
	}
	L := cfg.Logger().Named("ssh")
	ctx := pubd.WithSignalHandler(context.Background())
	return pubd.ListenAndServe(ctx, cfg.Addr, Server(L, fs, hostKey, cfg.ServerConfig))
}

func main() {
	if err := Main(osfs.New("/"), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
