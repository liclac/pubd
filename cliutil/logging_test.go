package cliutil

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestQuietVerbose(t *testing.T) {
	testdata := map[string]struct {
		Debug, Info, Warn, Error, DPanic, Panic, Fatal bool
	}{
		"":         {false, true, true, true, true, true, true},
		"-v":       {true, true, true, true, true, true, true},
		"-vv":      {true, true, true, true, true, true, true},
		"-q":       {false, false, true, true, true, true, true},
		"-qq":      {false, false, false, true, true, true, true},
		"-qqq":     {false, false, false, false, true, true, true},
		"-qqqq":    {false, false, false, false, false, true, true},
		"-qqqqq":   {false, false, false, false, false, false, true},
		"-qqqqqq":  {false, false, false, false, false, false, false},
		"-vq":      {true, false, true, true, true, true, true},
		"-vqq":     {true, false, false, true, true, true, true},
		"-vqqq":    {true, false, false, false, true, true, true},
		"-vqqqq":   {true, false, false, false, false, true, true},
		"-vqqqqq":  {true, false, false, false, false, false, true},
		"-vqqqqqq": {true, false, false, false, false, false, false},
	}
	for arg, tdata := range testdata {
		t.Run(`"`+arg+`"`, func(t *testing.T) {
			f := pflag.NewFlagSet("", 0)
			cfg := LogConfig{}
			cfg.Flags(f)
			require.NoError(t, f.Parse([]string{arg}))
			assert.Equal(t, tdata.Debug, cfg.Enabled(zapcore.DebugLevel), "debug")
			assert.Equal(t, tdata.Info, cfg.Enabled(zapcore.InfoLevel), "info")
			assert.Equal(t, tdata.Warn, cfg.Enabled(zapcore.WarnLevel), "warn")
			assert.Equal(t, tdata.Error, cfg.Enabled(zapcore.ErrorLevel), "error")
			assert.Equal(t, tdata.DPanic, cfg.Enabled(zapcore.DPanicLevel), "dpanic")
			assert.Equal(t, tdata.Panic, cfg.Enabled(zapcore.PanicLevel), "panic")
			assert.Equal(t, tdata.Fatal, cfg.Enabled(zapcore.FatalLevel), "fatal")
		})
	}
}
