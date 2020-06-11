package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/cliutil"
	"github.com/liclac/pubd/proto/httppub"
)

const Usage = `usage: pubd-http [path]`

type Config struct {
	Addr   string `toml:"addr"`
	Prefix string `toml:"prefix"`
	cliutil.FileSystemConfig
	cliutil.LogConfig
}

func Parse(fs billy.Filesystem, args []string) (Config, error) {
	cfg := Config{Addr: "localhost:8888", FileSystemConfig: cliutil.FileSystemDefaults()}
	return cfg, cliutil.Configure(&cfg, &cfg.Path, func(f *pflag.FlagSet) {
		f.StringVarP(&cfg.Addr, "addr", "a", cfg.Addr, "listen address")
		f.StringVarP(&cfg.Prefix, "prefix", "P", cfg.Prefix, "serve from a subdirectory")
		cfg.FileSystemConfig.Flags(f)
		cfg.LogConfig.Flags(f)
	}, Usage, args)
}

func (cfg *Config) Filesystem(fs billy.Filesystem) (http.FileSystem, error) {
	return cfg.FileSystemConfig.Build(fs)
}

func (cfg *Config) Handler(L *zap.Logger, fs http.FileSystem) http.Handler {
	return httppub.WithPrefix(cfg.Prefix,
		httppub.WithAccessLog(L.Named("access"),
			httppub.Handler(fs, httppub.SimpleIndex(), httppub.LogErrors(L.Named("req"))),
		))
}

func (cfg *Config) Server(L *zap.Logger, h http.Handler) pubd.Server {
	L = L.Named("server")
	return pubd.ServerFunc(func(ctx context.Context, l net.Listener) error {
		if ce := L.Check(zapcore.InfoLevel, "Running"); ce != nil {
			addr := fmt.Sprintf("http://%s%s/", l.Addr(), httppub.CleanPrefix(cfg.Prefix))
			ce.Write(zap.String("addr", addr))
		}
		return httppub.Serve(ctx, l, h)
	})
}

func Main(hostFS billy.Filesystem, args []string) error {
	cfg, err := Parse(hostFS, args)
	if err != nil {
		return err
	}
	fs, err := cfg.Filesystem(hostFS)
	if err != nil {
		return err
	}
	L := cfg.Logger().Named("http")
	ctx := pubd.WithSignalHandler(context.Background())
	return pubd.ListenAndServe(ctx, cfg.Addr, cfg.Server(L, cfg.Handler(L, fs)))
}

func main() {
	if err := Main(osfs.New("/"), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
