// Package logging provides a centralized, structured logging interface for the Cloud IDE backend.
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log severity levels.
type Level int

const (
	// DEBUG is for verbose development information.
	DEBUG Level = iota
	// INFO is for general operational messages.
	INFO
	// WARN is for non-critical issues that may need attention.
	WARN
	// ERROR is for failures that need immediate investigation.
	ERROR
	// FATAL is for unrecoverable errors that halt the process.
	FATAL
)

// String returns the human-readable name for a log level.
func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger is the centralized structured logger.
type Logger struct {
	mu       sync.Mutex
	level    Level
	output   *log.Logger
	fields   map[string]string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Default returns the singleton default logger instance.
func Default() *Logger {
	once.Do(func() {
		defaultLogger = New(os.Stdout, INFO)
	})
	return defaultLogger
}

// New creates a new Logger writing to the specified output at the given level.
// If w is nil, output is discarded.
func New(w io.Writer, level Level) *Logger {
	if w == nil {
		w = io.Discard
	}
	return &Logger{
		level:  level,
		output: log.New(w, "", 0),
		fields: make(map[string]string),
	}
}

// WithField returns a new logger with an additional structured field.
func (l *Logger) WithField(key, value string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]string, len(l.fields)+1)
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &Logger{
		level:  l.level,
		output: l.output,
		fields: newFields,
	}
}

// Debug logs a message at DEBUG level.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.log(DEBUG, msg, args...)
}

// Info logs a message at INFO level.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.log(INFO, msg, args...)
}

// Warn logs a message at WARN level.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.log(WARN, msg, args...)
}

// Error logs a message at ERROR level.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.log(ERROR, msg, args...)
}

// Fatal logs a message at FATAL level and exits the process.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.log(FATAL, msg, args...)
	os.Exit(1)
}

func (l *Logger) log(level Level, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	formatted := fmt.Sprintf(msg, args...)

	fieldStr := ""
	for k, v := range l.fields {
		fieldStr += fmt.Sprintf(" %s=%q", k, v)
	}

	l.output.Printf("%s [%s] %s%s", timestamp, level, formatted, fieldStr)
}
