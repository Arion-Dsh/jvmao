package jvmao

import (
	"io"
	"log/slog"
	"os"
	// "gopkg.in/natefinch/lumberjack.v2"
)

type Level int8

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	*slog.Logger
}

func DefaultLogger() *Logger {
	opts := LogOptions{
		Level:  LevelInfo,
		Stdout: true,
	}
	return NewLogger(opts)
}

// New returns a new logger.
func NewLogger(opt LogOptions) *Logger {
	return &Logger{
		slog.New(newHandler(opt)),
	}
}

func newHandler(opt LogOptions) slog.Handler {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}

	switch opt.Level {
	case LevelDebug:
		opts.Level = slog.LevelDebug
		opts.AddSource = true
	case LevelInfo:
		opts.Level = slog.LevelInfo
	case LevelWarn:
		opts.Level = slog.LevelWarn
	case LevelError:
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
	}

	ws := []io.Writer{os.Stdout}

	for _, w := range opt.MultiWriter {
		ws = append(ws, w)
	}

	/* if opt.logPath != "" { */
	/* fw := &lumberjack.Logger{ */
	/* Filename: opt.logPath, */
	/* MaxSize:  500, // megabytes */
	/* // MaxBackups: 3, */
	/* // MaxAge:   28,   //days */
	/* Compress: true, // disabled by default */
	/* } */
	/* if opt.logMaxAge > 0 { */
	/* fw.MaxAge = opt.logMaxAge */
	/* } */
	/* ws = append(ws, fw) */
	/* } */
	mw := io.MultiWriter(ws...)
	if opt.Format == "json" {
		return slog.NewJSONHandler(mw, opts)
	}
	return slog.NewTextHandler(mw, opts)

}

type LogOptions struct {
	Format      string // log type json or text
	Level       Level
	Stdout      bool
	MultiWriter []io.Writer
}
