package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// initOnce ensures zerolog globals are set exactly once,
// avoiding data races when multiple loggers are created concurrently.
var initOnce sync.Once

// Field is a structured key-value pair for log context.
type Field struct {
	Key   string
	Value interface{}
}

// F builds a structured log field.
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Err builds a field from an error.
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// Logger is the application-wide logging abstraction.
// It decouples business logic from any specific logging backend.
// Implementations must be safe for concurrent use.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
}

// Level represents the logging severity.
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// ParseLevel converts a string to a Level.
// Returns LevelInfo for unknown values.
func ParseLevel(s string) Level {
	switch s {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// zerologLogger implements Logger using the zerolog backend.
// Zerolog is zero-allocation, fast, and produces structured JSON output.
type zerologLogger struct {
	logger zerolog.Logger
}

// New creates a new Logger with the given level and output writer.
// If output is nil, it defaults to os.Stderr.
func New(level Level, output io.Writer) Logger {
	if output == nil {
		output = os.Stderr
	}

	zLevel := zerolog.InfoLevel
	switch level {
	case LevelDebug:
		zLevel = zerolog.DebugLevel
	case LevelInfo:
		zLevel = zerolog.InfoLevel
	case LevelWarn:
		zLevel = zerolog.WarnLevel
	case LevelError:
		zLevel = zerolog.ErrorLevel
	}

	initOnce.Do(func() {
		zerolog.TimeFieldFormat = time.RFC3339Nano
	})

	z := zerolog.New(output).
		Level(zLevel).
		With().
		Timestamp().
		Logger()

	return &zerologLogger{logger: z}
}

// NewDev creates a development-friendly (human-readable) logger.
// Uses zerolog's ConsoleWriter for colored output.
func NewDev(level Level) Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.Kitchen,
		NoColor:    false,
	}
	return New(level, output)
}

func (l *zerologLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug().Fields(convertFields(fields)).Msg(msg)
}

func (l *zerologLogger) Info(msg string, fields ...Field) {
	l.logger.Info().Fields(convertFields(fields)).Msg(msg)
}

func (l *zerologLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn().Fields(convertFields(fields)).Msg(msg)
}

func (l *zerologLogger) Error(msg string, fields ...Field) {
	l.logger.Error().Fields(convertFields(fields)).Msg(msg)
}

func (l *zerologLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal().Fields(convertFields(fields)).Msg(msg)
}

func (l *zerologLogger) With(fields ...Field) Logger {
	ctx := l.logger.With()
	for _, f := range fields {
		ctx = ctx.Interface(f.Key, f.Value)
	}
	child := ctx.Logger()
	return &zerologLogger{logger: child}
}

// NopLogger is a no-op implementation for testing.
type NopLogger struct{}

func (NopLogger) Debug(string, ...Field) {}
func (NopLogger) Info(string, ...Field)  {}
func (NopLogger) Warn(string, ...Field)  {}
func (NopLogger) Error(string, ...Field) {}
func (NopLogger) Fatal(string, ...Field) {}
func (NopLogger) With(...Field) Logger   { return NopLogger{} }

func convertFields(fields []Field) map[string]interface{} {
	m := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		m[f.Key] = f.Value
	}
	return m
}
