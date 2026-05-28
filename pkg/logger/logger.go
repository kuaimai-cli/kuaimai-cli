package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	mu      sync.Mutex
	verbose bool
	out     io.Writer = os.Stderr
)

// SetVerbose enables debug-level messages on stderr.
func SetVerbose(v bool) {
	mu.Lock()
	defer mu.Unlock()
	verbose = v
}

// Verbose reports whether debug logging is enabled.
func Verbose() bool {
	mu.Lock()
	defer mu.Unlock()
	return verbose
}

// Debug writes a debug line to stderr when verbose is on.
func Debug(format string, args ...any) {
	if !Verbose() {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	_, _ = fmt.Fprintf(out, "[debug] "+format+"\n", args...)
}

// Info writes an informational line to stderr.
func Info(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	_, _ = fmt.Fprintf(out, "[info] "+format+"\n", args...)
}

// Error writes an error line to stderr.
func Error(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	_, _ = fmt.Fprintf(out, "[error] "+format+"\n", args...)
}
