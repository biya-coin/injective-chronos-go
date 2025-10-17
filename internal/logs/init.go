package logs

import (
	"io"
	"os"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/logutil"
)

func init() {
	_ = os.MkdirAll("log", 0755)
	// 使用 SplitWriter，按前缀分流 API 与 CRON 日志
	sw, err := logutil.NewSplitWriter("log/api.log", "log/cron.log")
	if err == nil {
		w := logx.NewWriter(io.MultiWriter(os.Stdout, sw))
		logx.SetWriter(w)
	}
}
