package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	maxLogFiles = 7
	levelDebug  = "DEBUG"
	levelInfo   = "INFO"
	levelWarn   = "WARN"
	levelError  = "ERROR"
)

var (
	mu      sync.Mutex
	logFile *os.File
	logDay  string // current open log file date string YYYY-MM-DD
	logDir  string
)

// Init opens (or creates) the log file and rotates old files.
// logDirectory is typically %LOCALAPPDATA%\ChildMonitor\
func Init(logDirectory string) error {
	mu.Lock()
	defer mu.Unlock()

	logDir = logDirectory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	if err := rotateOldLogs(); err != nil {
		// Non-fatal: best-effort rotation.
		fmt.Fprintf(os.Stderr, "log rotation warning: %v\n", err)
	}

	return openLogFile()
}

// Close flushes and closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

func Debug(msg string, args ...any) { write(levelDebug, msg, args...) }
func Info(msg string, args ...any)  { write(levelInfo, msg, args...) }
func Warn(msg string, args ...any)  { write(levelWarn, msg, args...) }
func Error(msg string, args ...any) { write(levelError, msg, args...) }

func write(level, msg string, args ...any) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	today := now.Format("2006-01-02")

	// Roll to a new log file on day change.
	if logFile != nil && today != logDay {
		_ = logFile.Close()
		logFile = nil
		_ = rotateOldLogs()
		_ = openLogFile()
	}

	line := fmt.Sprintf("[%s] [%s] %s", now.Format("2006-01-02 15:04:05"), level, msg)
	if len(args) > 0 {
		line += " " + fmt.Sprint(args...)
	}
	line += "\n"

	writers := []io.Writer{os.Stderr}
	if logFile != nil {
		writers = append(writers, logFile)
	}
	for _, w := range writers {
		_, _ = fmt.Fprint(w, line)
	}
}

// openLogFile opens today's log file for appending. Must be called with mu held.
func openLogFile() error {
	today := time.Now().Format("2006-01-02")
	path := filepath.Join(logDir, fmt.Sprintf("app-%s.log", today))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logFile = f
	logDay = today
	return nil
}

// rotateOldLogs deletes log files beyond maxLogFiles. Must be called with mu held.
func rotateOldLogs() error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	var logs []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "app-") && strings.HasSuffix(e.Name(), ".log") {
			logs = append(logs, filepath.Join(logDir, e.Name()))
		}
	}

	// Sort ascending by name (date-based names sort correctly).
	sort.Strings(logs)

	for len(logs) >= maxLogFiles {
		if err := os.Remove(logs[0]); err != nil {
			return err
		}
		logs = logs[1:]
	}
	return nil
}
