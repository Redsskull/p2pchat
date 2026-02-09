package logger

import (
	"io"
	"log"
	"os"
)

var (
	// Global logger instances
	DebugLogger *log.Logger
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// Default: logs go to stderr
	DebugLogger = log.New(os.Stderr, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	InfoLogger = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// SetOutput redirects all logging to a specific writer
func SetOutput(w io.Writer) {
	DebugLogger.SetOutput(w)
	InfoLogger.SetOutput(w)
	ErrorLogger.SetOutput(w)
}

// Silent disables all logging
func Silent() {
	SetOutput(io.Discard)
}

// ToFile redirects logging to a file
func ToFile(filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	SetOutput(file)
	return nil
}

// Convenience functions
func Debug(format string, v ...any) {
	DebugLogger.Printf(format, v...)
}

func Info(format string, v ...any) {
	InfoLogger.Printf(format, v...)
}

func Error(format string, v ...any) {
	ErrorLogger.Printf(format, v...)
}
