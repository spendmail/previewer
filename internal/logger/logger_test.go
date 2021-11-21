package logger

import (
	"io/ioutil"
	"os"
	"testing"

	internalconfig "github.com/spendmail/otus_go_hw/hw12_13_14_15_calendar/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("logger", func(t *testing.T) {
		config, err := internalconfig.NewConfig("../../configs/calendar.toml")
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.CreateTemp("/tmp/", "")
		if err != nil {
			t.Fatal(err)
		}

		config.Logger.Level = "debug"
		config.Logger.File = f.Name()

		logger := New(config)

		logger.Debug("debug_message")
		logger.Info("info_message")
		logger.Warn("warn_message")
		logger.Error("error_message")

		b, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		content := string(b)

		require.Contains(t, content, "debug\tdebug_message", "Log doesn't contain string error")
		require.Contains(t, content, "info\tinfo_message", "Log doesn't contain string error")
		require.Contains(t, content, "warn\twarn_message", "Log doesn't contain string error")
		require.Contains(t, content, "error\terror_message", "Log doesn't contain string error")
	})
}
