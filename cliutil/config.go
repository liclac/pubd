package cliutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/pflag"
)

func Configure(
	out interface{}, posArg *string,
	registerFlags func(*pflag.FlagSet), usage string,
	args []string,
) error {
	flagSet := pflag.NewFlagSet(filepath.Base(args[0]), pflag.ContinueOnError)

	// Assumption: registerFlags uses String/StringVar/etc to populate `out` on flagSet.Parse().
	// Problem: If -C/--config is set, we want to parse a config file, but not override flags.
	// Solution: Start by parsing a substitute FlagSet with only that flag, and no error handling.
	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	flagSet.Usage = func() {}
	cfgFile := flagSet.StringP("config", "C", "", "load a config file")
	if flagSet.Parse(args) == nil && *cfgFile != "" {
		data, err := ioutil.ReadFile(*cfgFile)
		if err != nil {
			return fmt.Errorf("-C/--config: %w", err)
		}
		if err := toml.Unmarshal(data, out); err != nil {
			return fmt.Errorf("-C/--config: %w", err)
		}
	}

	// Now we can parse the actual flags, with error handling this time.
	flagSet.ParseErrorsWhitelist.UnknownFlags = false
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		fmt.Fprintln(os.Stderr, flagSet.FlagUsagesWrapped(80))
	}
	registerFlags(flagSet)
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	// Give posArg the first argument, if any. Refuse leftover arguments.
	posArgs := flagSet.Args()
	if len(posArgs) > 1 && posArg != nil {
		*posArg = posArgs[1]
		posArgs = posArgs[2:]
	}
	if len(posArgs) > 1 {
		flagSet.Usage()
		return fmt.Errorf("too many arguments")
	}

	return nil
}
