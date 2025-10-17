package logutil

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
)

// SplitWriter is an io.Writer that routes logs to different files based on a prefix.
// If the log line starts with "[CRON]" or "CRON|", it goes to cron file; otherwise to api file.
type SplitWriter struct {
	apiFile  *os.File
	cronFile *os.File
	mu       sync.Mutex
}

func NewSplitWriter(apiPath, cronPath string) (*SplitWriter, error) {
	if err := os.MkdirAll(filepath.Dir(apiPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cronPath), 0o755); err != nil {
		return nil, err
	}
	apiF, err := os.OpenFile(apiPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	cronF, err := os.OpenFile(cronPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = apiF.Close()
		return nil, err
	}
	return &SplitWriter{apiFile: apiF, cronFile: cronF}, nil
}

func (w *SplitWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if bytes.HasPrefix(p, []byte("[CRON]")) || bytes.HasPrefix(p, []byte("CRON|")) {
		return w.cronFile.Write(p)
	}
	return w.apiFile.Write(p)
}
