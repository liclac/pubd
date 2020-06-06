package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/cliutil"
	"github.com/liclac/pubd/proto/httppub"
)

const Usage = `usage: pubd-http [path]`

type Config struct {
	Quiet  bool   `toml:"quiet"`
	Addr   string `toml:"addr"`
	Prefix string `toml:"prefix"`
	cliutil.FileSystemConfig
}

func Parse(fs billy.Filesystem, args []string) (Config, error) {
	cfg := Config{Addr: "localhost:8888", FileSystemConfig: cliutil.FileSystemDefaults()}
	return cfg, cliutil.Configure(&cfg, &cfg.Path, func(f *pflag.FlagSet) {
		f.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, "don't print URL on startup")
		f.StringVarP(&cfg.Addr, "addr", "a", cfg.Addr, "listen address")
		f.StringVarP(&cfg.Prefix, "prefix", "P", cfg.Prefix, "serve from a subdirectory")
		cfg.FileSystemConfig.Flags(f)
	}, Usage, args)
}

func (cfg *Config) Filesystem(fs billy.Filesystem) (http.FileSystem, error) {
	return cfg.FileSystemConfig.Build(fs)
}

func (cfg *Config) Handler(fs http.FileSystem) http.Handler {
	return httppub.WithPrefix(cfg.Prefix,
		httppub.Handler(fs, httppub.SimpleIndex(), func(err error) {
			log.Printf("[ERR] http: %s", err)
		}))
}

func (cfg *Config) Listen() ([]net.Listener, error) {
	return pubd.Listen(cfg.Addr)
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
	h := cfg.Handler(fs)
	ls, err := cfg.Listen()
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(pubd.WithSignalHandler(context.Background()))
	for _, l := range ls {
		l := l
		g.Go(func() error {
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Running on: http://%s%s/\n", l.Addr(),
					httppub.CleanPrefix(cfg.Prefix))
			}
			return httppub.Serve(ctx, l, h)
		})
	}
	return g.Wait()
}

func main() {
	if err := Main(osfs.New("/"), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
