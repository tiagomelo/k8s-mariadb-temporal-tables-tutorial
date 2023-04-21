// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package handlers

import (
	"database/sql"
	"net/http"
	"os"

	v1 "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers/v1"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/middlewares/v1"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Db       *sql.DB
	Log      *zap.SugaredLogger
}

// NewAPIMux constructs a http.Handler with all application routes defined.
func NewAPIMux(cfg APIMuxConfig) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	var app *web.App

	if app == nil {
		app = web.NewApp(
			cfg.Shutdown,
			middlewares.Logger(cfg.Log),
			middlewares.Errors(cfg.Log),
		)
	}

	// Load the v1 routes.
	v1.Routes(app, v1.Config{
		Db:  cfg.Db,
		Log: cfg.Log,
	})

	return app
}
