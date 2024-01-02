package logger

import (
	"log/slog"
	"os"

	"github.com/dusted-go/logging/prettylog"
)

func Get(isProduction bool) *slog.Logger {
	if isProduction {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   false,
			ReplaceAttr: nil,
		}))
	} else {
		prettyHandler := prettylog.NewHandler(&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   false,
			ReplaceAttr: nil,
		})
		return slog.New(prettyHandler)
	}
}

func Field(key string, value any) slog.Attr {
	return slog.Any(key, value)
}
