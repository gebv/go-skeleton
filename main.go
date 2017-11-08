package main

import (
	"os"

	"github.com/gebv/go-skeleton/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	cli "gopkg.in/urfave/cli.v2"
)

var (
	VERSION = "dev"
	NAME    = "project_name"
	DEBUG   bool

	log *zap.SugaredLogger
)

func main() {
	app := &cli.App{}
	app.Name = NAME
	app.Version = VERSION
	app.Flags = []cli.Flag{
		// debug, info, warn, error
		&cli.StringFlag{Name: "loglevel", Value: "info", EnvVars: []string{"APP_LOGLEVEL"}, Usage: "log level: debug, info, warn, error"},

		&cli.BoolFlag{Name: "debug", EnvVars: []string{"APP_DEBUG"}, Destination: &DEBUG},
	}
	app.Before = func(c *cli.Context) (err error) {
		initAndSetLogLevel(c.String("loglevel"))
		return nil
	}
	app.Commands = []*cli.Command{}
	app.Run(os.Args)
}

func initAndSetLogLevel(strLevel string) {
	var level zapcore.Level
	switch strLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.DebugLevel
	}
	log = logger.New(level).Sugar()
	log.Infof("set log level %v", level)
}
