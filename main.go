package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/server"
)

var (
	VERSION string
)

// ServerConfig specifies the configuration for the metadata server
type ServerConfig struct {
	listen       string
	listenReload string
	enableXff    bool

	router     *mux.Router
	store      content.Store
	reloadChan chan os.Signal
}

func main() {
	app := cli.NewApp()
	app.Action = appMain
	app.Version = VERSION
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Debug",
		},
		cli.BoolFlag{
			Name:  "xff",
			Usage: "X-Forwarded-For header support",
		},
		cli.StringFlag{
			Name:  "listen",
			Value: ":80",
			Usage: "Address to listen to (TCP)",
		},
		cli.StringFlag{
			Name:   "access-key",
			EnvVar: "CATTLE_ACCESS_KEY",
			Usage:  "Rancher access key",
		},
		cli.StringFlag{
			Name:   "secret-key",
			EnvVar: "CATTLE_SECRET_KEY",
			Usage:  "Rancher secret key",
		},
		cli.StringFlag{
			Name:   "url",
			EnvVar: "CATTLE_URL",
			Usage:  "Rancher URL",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "",
			Usage: "Log file",
		},
	}

	app.Run(os.Args)
}

func appMain(ctx *cli.Context) error {
	if ctx.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logFile := ctx.GlobalString("log")
	if logFile != "" {
		if output, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
			logrus.Fatalf("Failed to log to file %s: %v", logFile, err)
		} else {
			logrus.SetOutput(output)
		}
	}

	opts := &client.ClientOpts{
		Url:       ctx.GlobalString("url"),
		AccessKey: ctx.GlobalString("access-key"),
		SecretKey: ctx.GlobalString("secret-key"),
	}

	s, err := server.New(opts,
		ctx.GlobalString("listen"),
		ctx.GlobalBool("xff"))

	if err != nil {
		return err
	}

	// Start the server
	return s.Start()
}
