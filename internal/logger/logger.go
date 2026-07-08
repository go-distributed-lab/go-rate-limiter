package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents log severity.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

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
	default:
		return "UNKNOWN"
	}
}

// Logger is a structured key=value logger backed by stdlib log.
type Logger struct {
	inner *log.Logger
	level Level
}

// New creates a Logger writing to w at the given minimum level.
func New(w io.Writer, level Level) *Logger {
	return &Logger{
		inner: log.New(w, "", 0),
		level: level,
	}
}

// Default returns a logger writing to stderr at INFO level.
func Default() *Logger {
	return New(os.Stderr, INFO)
}

// log emits a structured line: time=... level=... msg=... k=v ...
func (l *Logger) log(level Level, msg string, fields ...any) {
	if level < l.level {
		return
	}
	line := fmt.Sprintf("time=%s level=%s msg=%q",
		time.Now().Format(time.RFC3339),
		level.String(),
		msg,
	)
	for i := 0; i+1 < len(fields); i += 2 {
		line += fmt.Sprintf(" %v=%v", fields[i], fields[i+1])
	}
	l.inner.Println(line)
}

func (l *Logger) Info(msg string, fields ...any)  { l.log(INFO, msg, fields...) }
func (l *Logger) Warn(msg string, fields ...any)  { l.log(WARN, msg, fields...) }
func (l *Logger) Error(msg string, fields ...any) { l.log(ERROR, msg, fields...) }
func (l *Logger) Debug(msg string, fields ...any) { l.log(DEBUG, msg, fields...) }

// Discard returns a logger that drops all output. Used in tests/benchmarks.
func Discard() *Logger {
	return New(io.Discard, ERROR+1)
}
