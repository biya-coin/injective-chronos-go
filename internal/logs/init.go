package logs

import (
	"os"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/biya-coin/injective-chronos-go/internal/logutil"
)

func init() {
	_ = os.MkdirAll("log", 0755)
	// 使用 SplitWriter，按标记分流 API 与 CRON 日志；API 同步输出到 console
	sw, err := logutil.NewSplitWriter("log/api.log", "log/cron.log")
	if err == nil {
		// plain 格式；若 main 中已 SetUp，以 main 的设置为准
		_ = logx.SetUp(logx.LogConf{Encoding: "plain"})
		logx.SetWriter(logx.NewWriter(sw))
	}
}
