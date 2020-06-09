package cliutil

import (
	"os"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Common configuration for logging.
//
// Note that Quiet and Verbose are not opposites; setting both does not "cancel out",
// it will enable debug logging while also disabling info logging.
// This is done by implementing zapcore.LevelEnabler on LogConfig itself.
type LogConfig struct {
	Quiet   int `toml:"quiet"`   // Disable info/warn/error logging.
	Verbose int `toml:"verbose"` // Enable debug logging.
}

func (cfg *LogConfig) Flags(f *pflag.FlagSet) {
	f.CountVarP(&cfg.Quiet, "quiet", "q", "disable info/warn/error logging")
	f.CountVarP(&cfg.Verbose, "verbose", "v", "enable debug logging")
}

// zapcore.LevelEnabler
func (cfg LogConfig) Enabled(lvl zapcore.Level) bool {
	// Levels: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	if lvl < 0 {
		return cfg.Verbose >= int(-lvl) // Verbose=1 >= -Debug=-(-1)=1
	}
	return int(lvl) >= cfg.Quiet // Info(0) >= Quiet=1
}

func (cfg LogConfig) Logger() *zap.Logger {
	encCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "lvl",
		TimeKey:        "ts",
		NameKey:        "name",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	enc := zapcore.NewConsoleEncoder(encCfg)
	return zap.New(zapcore.NewCore(enc, os.Stderr, cfg))
}
