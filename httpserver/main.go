package main

import (
	"context"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"time"
	"turtorial.lendea.cn/common/constant"
	"turtorial.lendea.cn/common/logger"
	version2 "turtorial.lendea.cn/common/version"
	"turtorial.lendea.cn/server"
)

// Build variables for govvv
var (
	GitCommit string
	GitBranch string
	BuildTime string //ldflags
	Version   string
)

var (
	listenAddress = kingpin.Flag("web.listen-address",
		"Address to listen on for web interface and telemetry.").
		Default(":8080").Envar("WEB_LISTEN_ADDRESS").String()
	serviceName = kingpin.Flag("web.serviceName",
		"Service name.").
		Default(constant.APIName).Envar("WEB_SERVICE_NAME").String()
	version = kingpin.Flag("version", "project version.").
		Default("1.0.0").Envar("VERSION").String()
	showVersion = kingpin.Flag("v", "show version and exit").
			Default("false").Envar("V").Bool()
	logLevel = kingpin.Flag("log.level",
		"Sets the loglevel. Valid levels are debug,info, warn, error, dpanic, panic, fatal").
		Default("debug").Envar("LOG_LEVEL").String()
)

func main() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()
	if *showVersion {
		version2.ShowVersion(Version, GitCommit, GitBranch, BuildTime)
	}

	logger.Init(&logger.Config{
		ServiceName: *serviceName,
		LogLevel:    *logLevel,
	})

	// Create a context that is cancelled on SIGKILL or SIGINT.
	mainCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// create a http server
	httpServer := server.MakeHTTPServer(mainCtx, *version)
	httpServer.Addr = *listenAddress

	go func() {
		logger.For(mainCtx).Infof("starting https[%s] server...,version:[%s]", httpServer.Addr, *version)
		if err := httpServer.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.For(mainCtx).Info(err.Error())
			} else {
				logger.For(mainCtx).Fatal(err.Error())
			}
		}
	}()

	<-mainCtx.Done()

	// create a context for graceful http server shutdown
	shutdownHttpServer(mainCtx, httpServer)
}

//shutdownHttpServer
func shutdownHttpServer(mainCtx context.Context, httpServer *http.Server) {
	logger.For(mainCtx).Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.For(mainCtx).Error(err.Error())
	}
	logger.For(mainCtx).Info("Success shutdown server.")
}
