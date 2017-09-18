package main

import (
	"os"

	"context"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/rancher/go-rancher/v3"
	"github.com/rancher/metadata/content"
	"github.com/rancher/metadata/k8sproxy"
	"github.com/rancher/metadata/server"
	"golang.org/x/sync/errgroup"
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
			Value: "169.254.169.250:9346",
			Usage: "Address to listen to (TCP)",
		},
		cli.BoolFlag{
			Name:  "k8s-proxy",
			Usage: "Setup k8s api proxy on localhost",
		},
		cli.StringFlag{
			Name:  "k8s-proxy-http-listen",
			Value: "169.254.169.250:9347",
			Usage: "Address to listen to (HTTP)",
		},
		cli.StringFlag{
			Name:  "k8s-proxy-https-listen",
			Value: "169.254.169.250:9348",
			Usage: "Address to listen to (HTTPS)",
		},
		cli.StringFlag{
			Name:  "cert-file",
			Value: "/etc/kubernetes/ssl/cert.pem",
			Usage: "TLS cert for k8s proxy",
		},
		cli.StringFlag{
			Name:  "key-file",
			Value: "/etc/kubernetes/ssl/key.pem",
			Usage: "Private key for k8s proxy",
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

	group, _ := errgroup.WithContext(context.Background())
	group.Go(s.Start)

	if ctx.GlobalBool("k8s-proxy") {
		c, err := client.NewRancherClient(opts)
		if err != nil {
			return err
		}

		proxy, err := k8sproxy.New(c,
			ctx.GlobalString("k8s-proxy-http-listen"),
			ctx.GlobalString("k8s-proxy-https-listen"),
			ctx.GlobalString("cert-file"),
			ctx.GlobalString("key-file"))
		if err != nil {
			return err
		}

		group.Go(proxy.ListenAndServe)
	}

	// Start the server
	return group.Wait()
}
