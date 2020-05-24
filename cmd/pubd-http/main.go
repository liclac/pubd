package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	"github.com/liclac/pubd"
	"github.com/liclac/pubd/httppub"
)

var (
	fQuiet = pflag.BoolP("quiet", "q", false, "disable info logging")
	fAddr  = pflag.StringP("addr", "a", "127.0.0.1:8000", "listen address")
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
		log.Printf("Listening on: %s", l.Addr())
	}
	return httppub.Server{
		Root: fs,
	}.Serve(pubd.WithSignalHandler(context.Background()), l)
}

func main() {
	if err := Main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
