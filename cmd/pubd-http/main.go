package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/httppub"
)

var (
	fQuiet  = pflag.BoolP("quiet", "q", false, "don't print URL on startup")
	fAddr   = pflag.StringP("addr", "a", "localhost:8000", "listen address")
	fPrefix = pflag.StringP("prefix", "P", "", "serve from a subdirectory")
)

func Main() error {
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage:", filepath.Base(os.Args[0]), "[path]")
		pflag.PrintDefaults()
	}
	fsCfg := pubd.FileSystemFlags(pflag.CommandLine)
	pflag.Parse()

	fs, err := fsCfg.Build(pflag.Arg(0))
	if err != nil {
		return err
	}

	l, err := pubd.Listen(*fAddr)
	if err != nil {
		return err
	}
	if !*fQuiet {
		fmt.Fprintf(os.Stderr, "Running on: http://%s%s/\n", l.Addr(), httppub.CleanPrefix(*fPrefix))
	}
	return httppub.Serve(pubd.WithSignalHandler(context.Background()), l,
		httppub.WithPrefix(*fPrefix, httppub.Handler(fs)))
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
