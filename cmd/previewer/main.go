package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	internalapp "github.com/spendmail/previewer/internal/app"
	internalcache "github.com/spendmail/previewer/internal/cache"
	internalconfig "github.com/spendmail/previewer/internal/config"
	internallogger "github.com/spendmail/previewer/internal/logger"
	internalresizer "github.com/spendmail/previewer/internal/resizer"
	internalserver "github.com/spendmail/previewer/internal/server/http"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/previewer/previewer.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// Config initialization.
	config, err := internalconfig.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Logger initialization.
	logger, err := internallogger.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Cache initialization.
	cache, err := internalcache.New(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	// Application initialization.
	app, err := internalapp.New(logger, internalresizer.New(), cache)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP server initialization.
	server := internalserver.New(config, logger, app)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Locking until OS signal is sent or context cancel func is called.
		<-ctx.Done()

		// Stopping http server.
		stopHTTPCtx, stopHTTPCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer stopHTTPCancel()
		if err := server.Stop(stopHTTPCtx); err != nil {
			logger.Error(err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("starting http server...")

		// Locking over here until server is stopped.
		if err := server.Start(); err != nil {
			logger.Error(err.Error())
			cancel()
		}
	}()

	wg.Wait()
}
