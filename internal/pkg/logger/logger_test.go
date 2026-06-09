package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anomalyco/story/internal/pkg/logger"
)

func TestNew_OutputFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(logger.LevelInfo, &buf)

	log.Info("test message", logger.F("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain message, got: %s", output)
	}
	if !strings.Contains(output, "key") || !strings.Contains(output, "value") {
		t.Errorf("output should contain fields, got: %s", output)
	}
}

func TestLevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(logger.LevelError, &buf)

	log.Debug("should not appear")
	log.Info("should not appear either")
	log.Error("should appear")

	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Error("debug/info messages should be filtered at Error level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("error messages should appear at Error level")
	}
}

func TestWith_Fields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	log := logger.New(logger.LevelInfo, &buf)

	ctxLog := log.With(logger.F("request_id", "abc-123"))
	ctxLog.Info("handling request")

	output := buf.String()
	if !strings.Contains(output, "abc-123") {
		t.Errorf("contextual fields should appear in output, got: %s", output)
	}
}

func TestNopLogger(t *testing.T) {
	t.Parallel()

	log := logger.NopLogger{}
	// These should not panic
	log.Debug("test")
	log.Info("test", logger.F("k", "v"))
	log.With(logger.F("k", "v")).Info("test")
	log.Error("test")
	log.Fatal("test")
	log.Warn("test")
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
		{"unknown", "info"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := logger.ParseLevel(tt.input)
			if string(got) != tt.want {
				t.Errorf("ParseLevel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewDev_Output(t *testing.T) {
	t.Parallel()

	log := logger.NewDev(logger.LevelDebug)
	// Should not panic — visual verification only
	log.Info("dev mode test", logger.F("component", "test"))
}
