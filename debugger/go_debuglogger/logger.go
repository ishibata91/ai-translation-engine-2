package go_debuglogger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger is a disposable debugger logger.
// Importers should keep calls isolated so the import and usage can be deleted in one sweep.
type Logger struct {
	mu     sync.Mutex
	file   *os.File
	path   string
	module string
}

// Event is one JSONL row emitted by the debugger logger.
type Event struct {
	Timestamp string         `json:"timestamp"`
	Level     string         `json:"level"`
	Module    string         `json:"module"`
	Stage     string         `json:"stage"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// New creates a logger that appends JSONL rows to the provided path.
func New(path string, module string) (*Logger, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "" || cleanPath == "." {
		return nil, fmt.Errorf("debugger log path is required")
	}
	if module == "" {
		module = "debugger"
	}
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o755); err != nil {
		return nil, fmt.Errorf("create debugger log dir: %w", err)
	}
	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open debugger log file: %w", err)
	}
	return &Logger{file: file, path: cleanPath, module: module}, nil
}

// NewFromEnv creates a logger from AITE2_DEBUGGER_EVENTS_PATH.
func NewFromEnv(module string) (*Logger, error) {
	path := os.Getenv("AITE2_DEBUGGER_EVENTS_PATH")
	if path == "" {
		return nil, fmt.Errorf("AITE2_DEBUGGER_EVENTS_PATH is not set")
	}
	return New(path, module)
}

func (l *Logger) Path() string {
	if l == nil {
		return ""
	}
	return l.path
}

func (l *Logger) Close() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *Logger) Info(ctx context.Context, stage string, message string, fields map[string]any) error {
	return l.write(ctx, "info", stage, message, fields)
}

func (l *Logger) Warn(ctx context.Context, stage string, message string, fields map[string]any) error {
	return l.write(ctx, "warn", stage, message, fields)
}

func (l *Logger) Error(ctx context.Context, stage string, message string, fields map[string]any) error {
	return l.write(ctx, "error", stage, message, fields)
}

func (l *Logger) write(ctx context.Context, level string, stage string, message string, fields map[string]any) error {
	if l == nil || l.file == nil {
		return fmt.Errorf("debugger logger is not initialized")
	}
	payload := cloneFields(fields)
	if ctx != nil {
		payload["context_present"] = true
	}
	entry := Event{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Level:     level,
		Module:    l.module,
		Stage:     stage,
		Message:   message,
		Fields:    payload,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal debugger event: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.file.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write debugger event: %w", err)
	}
	return nil
}

func cloneFields(fields map[string]any) map[string]any {
	if len(fields) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(fields))
	for key, value := range fields {
		cloned[key] = value
	}
	return cloned
}
