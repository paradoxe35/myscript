package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type FileLogger struct {
	logger *log.Logger
	file   *os.File
}

func NewFileLogger(logDir string) (*FileLogger, error) {
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
	logger := log.New(multiWriter, "", log.Ldate|log.Ltime)

	return &FileLogger{
		logger: logger,
		file:   file,
	}, nil
}

func (l *FileLogger) log(level, message string) {
	_, file, line, _ := runtime.Caller(2)
	l.logger.Printf("[%s] %s:%d: %s", level, filepath.Base(file), line, message)
}

func (l *FileLogger) Print(message string) {
	l.log("PRINT", message)
}

func (l *FileLogger) Trace(message string) {
	l.log("TRACE", message)
}

func (l *FileLogger) Debug(message string) {
	l.log("DEBUG", message)
}

func (l *FileLogger) Info(message string) {
	l.log("INFO", message)
}

func (l *FileLogger) Warning(message string) {
	l.log("WARNING", message)
}

func (l *FileLogger) Error(message string) {
	l.log("ERROR", message)
}

func (l *FileLogger) Fatal(message string) {
	l.log("FATAL", message)
	os.Exit(1)
}

func (l *FileLogger) Close() error {
	return l.file.Close()
}
