package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, INFO)

	log.Debug("should not appear")
	if buf.Len() > 0 {
		t.Error("DEBUG message should not appear at INFO level")
	}

	log.Info("info message")
	if !strings.Contains(buf.String(), "[INFO]") {
		t.Error("expected INFO level in output")
	}
	if !strings.Contains(buf.String(), "info message") {
		t.Error("expected message text in output")
	}
}

func TestLoggerWithField(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, INFO)

	fieldLog := log.WithField("component", "test")
	fieldLog.Info("test message")

	output := buf.String()
	if !strings.Contains(output, `component="test"`) {
		t.Errorf("expected field in output, got: %s", output)
	}
}

func TestLoggerNilWriter(t *testing.T) {
	log := New(nil, INFO)

	// Should not panic.
	log.Info("test message")
	log.Error("error message")
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.expected)
		}
	}
}

func TestDefaultLogger(t *testing.T) {
	log := Default()
	if log == nil {
		t.Fatal("expected non-nil default logger")
	}

	// Should return same instance.
	log2 := Default()
	if log != log2 {
		t.Error("expected same singleton instance")
	}
}
