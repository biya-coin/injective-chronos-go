package logs

import (
	"os"

	"github.com/zeromicro/go-zero/core/logx"
)

func init() {
	_ = os.MkdirAll("log", 0755)
	f, err := os.OpenFile("log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		w := logx.NewWriter(f)
		logx.SetWriter(w)
		// Intentionally not closing file; writer holds it for process lifetime.
	}
}
