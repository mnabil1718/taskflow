package logger

import (
	"log/slog"
	"os"
)

func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if env == "development" {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
