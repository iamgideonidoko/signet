package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"sync"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[Level]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type Logger struct {
	level  Level
	out    io.Writer
	mu     sync.Mutex
	fields map[string]any
}

type LogEntry struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
	Caller    string         `json:"caller,omitempty"`
}

var std *Logger

func init() {
	std = New(INFO, os.Stdout)
}

func New(level Level, out io.Writer) *Logger {
	return &Logger{
		level:  level,
		out:    out,
		fields: make(map[string]any),
	}
}

func SetLevel(level Level) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

func (l *Logger) WithField(key string, value any) *Logger {
	newLogger := &Logger{
		level:  l.level,
		out:    l.out,
		fields: make(map[string]any),
	}
	maps.Copy(newLogger.fields, l.fields)
	newLogger.fields[key] = value
	return newLogger
}

func (l *Logger) log(level Level, msg string, fields map[string]any) {
	if level < l.level {
		return
	}

	l.mu.Lock()

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     levelNames[level],
		Message:   msg,
		Fields:    make(map[string]any),
	}

	maps.Copy(entry.Fields, l.fields)
	maps.Copy(entry.Fields, fields)

	if level >= ERROR {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			entry.Caller = fmt.Sprintf("%s:%d", file, line)
		}
	}

	data, _ := json.Marshal(entry)
	_, _ = fmt.Fprintf(l.out, "%s\n", data)

	if level == FATAL {
		l.mu.Unlock() // Explicitly unlock before exit
		os.Exit(1)
	}

	l.mu.Unlock()
}

func (l *Logger) Info(msg string, fields ...map[string]any) {
	f := mergeFields(fields...)
	l.log(INFO, msg, f)
}

func (l *Logger) Warn(msg string, fields ...map[string]any) {
	f := mergeFields(fields...)
	l.log(WARN, msg, f)
}

func (l *Logger) Error(msg string, fields ...map[string]any) {
	f := mergeFields(fields...)
	l.log(ERROR, msg, f)
}

func Info(msg string, fields ...map[string]any) {
	std.Info(msg, fields...)
}

func Warn(msg string, fields ...map[string]any) {
	std.Warn(msg, fields...)
}

func Error(msg string, fields ...map[string]any) {
	std.Error(msg, fields...)
}

func WithField(key string, value any) *Logger {
	return std.WithField(key, value)
}

func mergeFields(fields ...map[string]any) map[string]any {
	result := make(map[string]any)
	for _, f := range fields {
		maps.Copy(result, f)
	}
	return result
}

func ParseLevel(level string) Level {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}
