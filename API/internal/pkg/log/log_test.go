package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	l := New(InfoLevel, &buf)

	l.Info("test message")
	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO in output, got: %s", out)
	}
	if !strings.Contains(out, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", out)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(InfoLevel, &buf)

	l.Info("message", map[string]any{"user": "alice", "count": 42})
	out := buf.String()
	if !strings.Contains(out, "user=alice") {
		t.Errorf("expected user=alice in output, got: %s", out)
	}
	if !strings.Contains(out, "count=42") {
		t.Errorf("expected count=42 in output, got: %s", out)
	}
}

func TestLoggerLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(WarnLevel, &buf)

	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")

	out := buf.String()
	if strings.Contains(out, "debug message") {
		t.Errorf("debug message should not be logged at WarnLevel")
	}
	if strings.Contains(out, "info message") {
		t.Errorf("info message should not be logged at WarnLevel")
	}
	if !strings.Contains(out, "warn message") {
		t.Errorf("warn message should be logged at WarnLevel")
	}
}

func TestWithField(t *testing.T) {
	var buf bytes.Buffer
	l := New(InfoLevel, &buf).WithField("request_id", "123")

	l.Info("test")
	out := buf.String()
	if !strings.Contains(out, "request_id=123") {
		t.Errorf("expected request_id=123 in output, got: %s", out)
	}
}
