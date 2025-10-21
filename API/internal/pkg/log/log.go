package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a simple structured logger.
type Logger struct {
	level  Level
	output io.Writer
	fields map[string]any
}

// New creates a new Logger with the given minimum level and output writer.
func New(level Level, output io.Writer) *Logger {
	if output == nil {
		output = os.Stderr
	}
	return &Logger{
		level:  level,
		output: output,
		fields: make(map[string]any),
	}
}

// Default returns a logger configured from environment variables.
// LOG_LEVEL can be: debug, info, warn, error (default: info)
func Default() *Logger {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	level := InfoLevel
	switch levelStr {
	case "debug":
		level = DebugLevel
	case "warn", "warning":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	}
	return New(level, os.Stderr)
}

// WithFields returns a new Logger with the given fields merged in.
func (l *Logger) WithFields(fields map[string]any) *Logger {
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{
		level:  l.level,
		output: l.output,
		fields: merged,
	}
}

// WithField returns a new Logger with a single field added.
func (l *Logger) WithField(key string, value any) *Logger {
	return l.WithFields(map[string]any{key: value})
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}

	// Build field string
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}

	var parts []string
	for k, v := range merged {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	fieldStr := ""
	if len(parts) > 0 {
		fieldStr = " " + strings.Join(parts, " ")
	}

	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s: %s%s\n", timestamp, level.String(), msg, fieldStr)
	l.output.Write([]byte(line))
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...map[string]any) {
	f := make(map[string]any)
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DebugLevel, msg, f)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...map[string]any) {
	f := make(map[string]any)
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(InfoLevel, msg, f)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...map[string]any) {
	f := make(map[string]any)
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WarnLevel, msg, f)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...map[string]any) {
	f := make(map[string]any)
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ErrorLevel, msg, f)
}

// Errorf logs a formatted error message.
func (l *Logger) Errorf(format string, args ...any) {
	l.Error(fmt.Sprintf(format, args...))
}

// global default logger for convenience
var defaultLogger = Default()

// SetDefault sets the global default logger.
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Debug logs using the default logger.
func Debug(msg string, fields ...map[string]any) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs using the default logger.
func Info(msg string, fields ...map[string]any) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs using the default logger.
func Warn(msg string, fields ...map[string]any) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs using the default logger.
func Error(msg string, fields ...map[string]any) {
	defaultLogger.Error(msg, fields...)
}

// Errorf logs a formatted error using the default logger.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// StdLogger returns a stdlib log.Logger that writes to the given logger at Info level.
func StdLogger(l *Logger) *log.Logger {
	return log.New(l.output, "[INFO] ", log.LstdFlags)
}
