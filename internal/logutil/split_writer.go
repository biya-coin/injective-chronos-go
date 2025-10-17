package logutil

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// SplitWriter is an io.Writer that routes logs to different files based on a prefix.
// If the log line starts with "[CRON]" or "CRON|", it goes to cron file; otherwise to api file.
type SplitWriter struct {
	apiFile  *os.File
	cronFile *os.File
	console  io.Writer
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
	return &SplitWriter{apiFile: apiF, cronFile: cronF, console: os.Stdout}, nil
}

func (w *SplitWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	// logx 会在每行前增加时间/级别等前缀，因此这里用 Contains 匹配标记
	if bytes.Contains(p, []byte("[CRON]")) || bytes.Contains(p, []byte("CRON|")) {
		// 只写入 cron 文件
		return w.cronFile.Write(p)
	}
	// API 日志：写文件并镜像到 console
	if _, err := w.apiFile.Write(p); err != nil {
		return 0, err
	}
	if w.console != nil {
		_, _ = w.console.Write(p)
	}
	return len(p), nil
}
