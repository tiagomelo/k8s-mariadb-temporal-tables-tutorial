// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/config"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/logger"
	"go.uber.org/zap"
)

const twenty = time.Second * 20

func handleShutdown(log *zap.SugaredLogger, shutdown chan os.Signal,
	serverErrors chan error, api *http.Server) error {
	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), twenty)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}
	return nil
}

func run(log *zap.SugaredLogger) error {
	// =========================================================================
	// App Starting

	log.Infow("starting service")
	defer log.Infow("shutdown complete")

	// =========================================================================
	// Reading config

	log.Infow("startup", "status", "reading config")
	cfg, err := config.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	// =========================================================================
	// Database support
	db, err := db.ConnectToMariaDb(cfg.MariaDbUser, cfg.MariaDbPassword, cfg.MariaDbHost, cfg.MariaDbPort, cfg.MariaDbDatabase)

	log.Infow("startup", "status", "initializing database support")
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing V1 API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct the mux for the API calls.
	apiMux := handlers.NewAPIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Db:       db,
		Log:      log,
	})

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:     ":8080",
		Handler:  apiMux,
		ErrorLog: zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	if err := handleShutdown(log, shutdown, serverErrors, &api); err != nil {
		return err
	}

	return nil
}

func main() {
	// Construct the application logger.
	log, err := logger.New("DOCKER-MARIADB-TEMPORAL-TABLES-TUTORIAL")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Println("error calling log.Sync():", err)
			os.Exit(1)
		}
	}()
	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}
