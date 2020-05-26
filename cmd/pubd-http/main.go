package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/cmd"
	"github.com/liclac/pubd/httppub"
)

const Usage = `usage: pubd-http [path]`

type Config struct {
	Quiet  bool   `toml:"quiet"`
	Addr   string `toml:"addr"`
	Prefix string `toml:"prefix"`
	pubd.FileSystemConfig
}

func (cfg *Config) Flags(f *pflag.FlagSet) {
	f.BoolVarP(&cfg.Quiet, "quiet", "q", cfg.Quiet, "don't print URL on startup")
	f.StringVarP(&cfg.Addr, "addr", "a", cfg.Addr, "listen address")
	f.StringVarP(&cfg.Prefix, "prefix", "P", cfg.Prefix, "serve from a subdirectory")
	cfg.FileSystemConfig.Flags(f)
}

func Main() error {
	cfg := Config{Addr: "localhost:8080"}
	if err := cmd.Configure(&cfg, &cfg.Path, cfg.Flags, Usage, os.Args); err != nil {
		return err
	}
	fs, err := cfg.FileSystemConfig.Build()
	if err != nil {
		return err
	}
	l, err := pubd.Listen(cfg.Addr)
	if err != nil {
		return err
	}
	if !cfg.Quiet {
		fmt.Fprintf(os.Stderr, "Running on: http://%s%s/\n", l.Addr(),
			httppub.CleanPrefix(cfg.Prefix))
	}
	return httppub.Serve(pubd.WithSignalHandler(context.Background()), l,
		httppub.WithPrefix(cfg.Prefix, httppub.Handler(fs)))
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
