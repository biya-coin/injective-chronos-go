package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/biya-coin/injective-chronos-go/internal/config"
	"github.com/biya-coin/injective-chronos-go/internal/handler"
	_ "github.com/biya-coin/injective-chronos-go/internal/logs"
	"github.com/biya-coin/injective-chronos-go/internal/svc"
	"github.com/biya-coin/injective-chronos-go/internal/task"
)

// Config file selection: flag -f overrides; otherwise use ENV (default "dev").
var configFile = flag.String("f", "", "the config file")

func main() {
	flag.Parse()
	// initialize logging via side-effect import

	var c config.Config
	// resolve config path based on -f flag or ENV
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	cfgPath := *configFile
	if cfgPath == "" {
		cfgPath = fmt.Sprintf("etc/config.%s.yaml", env)
	}
	conf.MustLoad(cfgPath, &c)

	logx.Infof("env=%s config=%s", env, cfgPath)
	logx.Infof("starting %s on %s:%d", c.Name, c.Host, c.Port)

	ctx := svc.NewServiceContext(c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)
	server.AddRoutes([]rest.Route{{
		Method: "GET",
		Path:   "/healthz",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			handler.HealthHandler(w, r)
		},
	}})

	// start cron
	task.StartCron(ctx)

	fmt.Printf("listening on %s:%d\n", c.Host, c.Port)
	server.Start()
}
