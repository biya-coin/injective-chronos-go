package logs

import (
	"io"
	"os"

	"github.com/zeromicro/go-zero/core/logx"
)

func init() {
	_ = os.MkdirAll("log", 0755)
	f, err := os.OpenFile("log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		// add out to console and file
		w := logx.NewWriter(io.MultiWriter(os.Stdout, f))
		logx.SetWriter(w)
	}
}
