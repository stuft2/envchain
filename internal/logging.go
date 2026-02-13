package internal

import (
	"io"
	"log"
	"sync"
)

var (
	loggerMu sync.RWMutex
	logger   = log.New(io.Discard, "", 0)
)

// SetLogger replaces the logger used for debug output. Pass nil to disable logging.
func SetLogger(l *log.Logger) {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if l == nil {
		logger = log.New(io.Discard, "", 0)
		return
	}
	logger = l
}

// Debugf writes a formatted message to the configured logger.
func Debugf(format string, args ...any) {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if logger == nil {
		return
	}
	logger.Printf(format, args...)
}
