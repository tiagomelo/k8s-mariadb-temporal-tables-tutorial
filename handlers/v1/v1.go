// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package v1

import (
	"database/sql"
	"net/http"

	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers/v1/db/timestamp"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers/v1/employees"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers/v1/employees/history"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Db  *sql.DB
	Log *zap.SugaredLogger
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"
	eh := employees.Handlers{Db: cfg.Db}
	ehh := history.Handlers{Db: cfg.Db}
	th := timestamp.Handlers{Db: cfg.Db}
	app.Handle(http.MethodGet, version, "/employees", eh.GetAll)
	app.Handle(http.MethodGet, version, "/employee/:id", eh.GetById)
	app.Handle(http.MethodPost, version, "/employee", eh.Create)
	app.Handle(http.MethodPut, version, "/employee/:id", eh.Update)
	app.Handle(http.MethodDelete, version, "/employee/:id", eh.Delete)
	app.Handle(http.MethodGet, version, "/employee/:id/history/all", ehh.GetAll)
	app.Handle(http.MethodGet, version, "/employee/:id/history/:timestamp", ehh.AtPointInTime)
	app.Handle(http.MethodGet, version, "/employee/:id/history/:startTimestamp/:endTimestamp", ehh.BetweenDates)
	app.Handle(http.MethodGet, version, "/db/timestamp/advance", th.AdvanceTimestamp)
	app.Handle(http.MethodGet, version, "/db/timestamp/default", th.SetDefaultTimestamp)
}
