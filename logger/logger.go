package logger

import (
	"io"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/echosonusharma/image-uploader-service/config"
)

// to close log file in main func
var LogFile *os.File

var Log *slog.Logger

func Init() error {
	file, err := os.OpenFile("./log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	LogFile = file

	loggingLevel := new(slog.LevelVar)
	writer := io.MultiWriter(LogFile, os.Stdout)

	switch config.Cfg.LOG_LEVEL {
	case "debug":
		loggingLevel.Set(slog.LevelDebug)
	case "info":
		loggingLevel.Set(slog.LevelInfo)
	case "warn":
		loggingLevel.Set(slog.LevelWarn)
	case "error":
		loggingLevel.Set(slog.LevelError)
	default:
		loggingLevel.Set(slog.LevelDebug)
	}

	replacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
			source.Function = filepath.Base(source.Function)
		}
		return a
	}

	Log = slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: loggingLevel, AddSource: true, ReplaceAttr: replacer}))

	return nil
}
