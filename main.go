/**
 *
 *    ____          __
 *   / __/__ ____ _/ /__
 *  / _// _ `/ _ `/ / -_)
 * /___/\_,_/\_, /_/\__/
 *         /___/
 *
 *
 * generate by http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=Eagle
 */
package main

import (
	"demo/internal/model"
	"demo/internal/server/websocket"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	eagle "github.com/go-eagle/eagle/pkg/app"
	"github.com/go-eagle/eagle/pkg/config"
	logger "github.com/go-eagle/eagle/pkg/log"
	"github.com/go-eagle/eagle/pkg/transport/grpc"
	v "github.com/go-eagle/eagle/pkg/version"
	"github.com/spf13/pflag"
	_ "go.uber.org/automaxprocs"

	"demo/internal/server"
)

var (
	cfgDir  = pflag.StringP("config dir", "c", "config", "config path.")
	env     = pflag.StringP("env name", "e", "", "env var name.")
	version = pflag.BoolP("version", "v", false, "show version info.")
)

// @title eagle docs api
// @version 1.0
// @description eagle demo

// @host localhost:8080
// @BasePath /v1
func main() {
	pflag.Parse()
	if *version {
		ver := v.Get()
		marshaled, err := json.MarshalIndent(&ver, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(marshaled))
		return
	}

	// init config
	c := config.New(*cfgDir, config.WithEnv(*env))
	var cfg eagle.Config
	if err := c.Load("app", &cfg); err != nil {
		panic(err)
	}
	// set global
	eagle.Conf = &cfg

	// -------------- init resource -------------
	logger.Init()

	model.Init()

	gin.SetMode(cfg.Mode)

	// init pprof server
	go func() {
		fmt.Printf("Listening and serving PProf HTTP on %s\n", cfg.PprofPort)
		if err := http.ListenAndServe(cfg.PprofPort, http.DefaultServeMux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen ListenAndServe for PProf, err: %s", err.Error())
		}
	}()

	// start app
	app, cleanup, err := InitApp(&cfg, &cfg.GRPC)
	go websocket.NewWsServer()
	defer cleanup()
	if err != nil {
		panic(err)
	}
	//var taskCfg tasks.Config
	//if err := c.Load("cronjob", &taskCfg); err != nil {
	//	panic(err)
	//}
	//cronJobService := service.NewCronJobService()
	//server.NewCronJobServer(&taskCfg, cronJobService)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func newApp(cfg *eagle.Config, gs *grpc.Server) *eagle.App {
	return eagle.New(
		eagle.WithName(cfg.Name),
		eagle.WithVersion(cfg.Version),
		eagle.WithLogger(logger.GetLogger()),
		eagle.WithServer(
			// init HTTP server
			server.NewHTTPServer(&cfg.HTTP),
			// init websocketServer
			server.NewWSServer(&cfg.HTTP),
			// init gRPC server
			gs,
		),
	)
}
