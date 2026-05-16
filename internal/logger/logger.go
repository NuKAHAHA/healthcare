package logger

import (
	"encoding/json"
	"log"
	"time"
)

// LogLevel represents log severity level
type LogLevel string

const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

// Logger provides structured JSON logging
type Logger struct {
	level LogLevel
}

// New creates a new logger
func New(level string) *Logger {
	logLevel := InfoLevel
	switch level {
	case "debug":
		logLevel = DebugLevel
	case "info":
		logLevel = InfoLevel
	case "warn":
		logLevel = WarnLevel
	case "error":
		logLevel = ErrorLevel
	}
	return &Logger{level: logLevel}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	UserID    interface{} `json:"userId,omitempty"`
	Action    string      `json:"action,omitempty"`
	Details   string      `json:"details,omitempty"`
}

// log is the internal logging function
func (l *Logger) log(level LogLevel, message, action, details string, userID interface{}) {
	// Only log if level is >= configured level
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   message,
		Action:    action,
		Details:   details,
		UserID:    userID,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	log.Println(string(data))
}

// Debug logs at debug level
func (l *Logger) Debug(message string) {
	l.log(DebugLevel, message, "", "", nil)
}

// Info logs at info level
func (l *Logger) Info(message string) {
	l.log(InfoLevel, message, "", "", nil)
}

// InfoWithContext logs at info level with context
func (l *Logger) InfoWithContext(message, action string, userID int64, details string) {
	l.log(InfoLevel, message, action, details, userID)
}

// Warn logs at warn level
func (l *Logger) Warn(message string) {
	l.log(WarnLevel, message, "", "", nil)
}

// WarnWithContext logs at warn level with context
func (l *Logger) WarnWithContext(message, action string, userID int64, details string) {
	l.log(WarnLevel, message, action, details, userID)
}

// Error logs at error level
func (l *Logger) Error(message string) {
	l.log(ErrorLevel, message, "", "", nil)
}

// ErrorWithContext logs at error level with context
func (l *Logger) ErrorWithContext(message, action string, userID int64, details string) {
	l.log(ErrorLevel, message, action, details, userID)
}

// shouldLog determines if a message at level should be logged
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DebugLevel: 0,
		InfoLevel:  1,
		WarnLevel:  2,
		ErrorLevel: 3,
	}
	return levels[level] >= levels[l.level]
}
