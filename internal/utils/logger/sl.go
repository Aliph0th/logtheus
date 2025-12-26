package sl

import (
	"log/slog"
	"logtheus/internal/consts"
	"os"
	"time"

	"github.com/Marlliton/slogpretty"
)

func SetupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case consts.DEVELOPMENT:
		logger = slog.New(slogpretty.New(os.Stdout, &slogpretty.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
			Colorful:   true,
		}))
	case consts.PRODUCTION:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return logger
}

func Error(error error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(error.Error()),
	}
}
