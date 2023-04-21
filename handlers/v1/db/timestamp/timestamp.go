// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package timestamp

import (
	"context"
	"database/sql"
	"net/http"

	mariaDb "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	v1Web "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web/v1"
)

type Handlers struct {
	Db *sql.DB
}

// AdvanceTimestamp advances timestamp in database. It is used mainly for demonstration
// purposes.
func (h Handlers) AdvanceTimestamp(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := mariaDb.AdvanceMariaDbTimestamp(ctx, h.Db); err != nil {
		return v1Web.NewRequestError(err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// SetDefaultTimestamp sets the database timestamp back to default, which is,
// the current date.
func (h Handlers) SetDefaultTimestamp(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := mariaDb.SetDefaultMariaDbTimestamp(ctx, h.Db); err != nil {
		return v1Web.NewRequestError(err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
