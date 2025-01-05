package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type FileLogger struct {
	logger *slog.Logger
	file   *os.File
}

type CustomLogger struct {
	Slog   *slog.Logger
	Logger *FileLogger
}

func NewFileLogger(homeDir string) (*CustomLogger, error) {
	logDir := filepath.Join(homeDir, "logs")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logDir, fmt.Sprintf("app-%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)
	opts := &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	}
	handler := slog.NewTextHandler(multiWriter, opts)
	logger := slog.New(handler)

	fileLogger := &FileLogger{
		logger: logger,
		file:   file,
	}

	return &CustomLogger{
		Slog:   logger,
		Logger: fileLogger,
	}, nil
}

func (l *FileLogger) log(level slog.Level, message string) {
	_, file, line, _ := runtime.Caller(2)
	attrs := []slog.Attr{
		slog.String("file", filepath.Base(file)),
		slog.Int("line", line),
	}

	l.logger.LogAttrs(context.TODO(), level, message, attrs...)
}

func (l *FileLogger) Print(message string) {
	l.log(slog.LevelInfo, message)
}

func (l *FileLogger) Trace(message string) {
	l.log(slog.LevelDebug-1, message) // Trace is typically lower than Debug
}

func (l *FileLogger) Debug(message string) {
	l.log(slog.LevelDebug, message)
}

func (l *FileLogger) Info(message string) {
	l.log(slog.LevelInfo, message)
}

func (l *FileLogger) Warning(message string) {
	l.log(slog.LevelWarn, message)
}

func (l *FileLogger) Error(message string) {
	l.log(slog.LevelError, message)
}

func (l *FileLogger) Fatal(message string) {
	l.log(slog.LevelError, message) // Using Error level before exiting
	os.Exit(1)
}

func (l *FileLogger) Close() error {
	return l.file.Close()
}
