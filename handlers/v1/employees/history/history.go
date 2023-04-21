// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package history

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	v1Web "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web/v1"
)

type Handlers struct {
	Db *sql.DB
}

// GetAll returns all historical data for a given employee.
func (h Handlers) GetAll(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	employeeHistory, err := employees.GetAllHistory(ctx, h.Db, uint(id))
	if err != nil {
		return fmt.Errorf("ID[%d]: %w", id, err)
	}
	return web.Respond(ctx, w, employeeHistory, http.StatusOK)
}

// AtPointInTime returns historical data for a given employee at a point in time.
func (h Handlers) AtPointInTime(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	timestampParam := web.Param(r, "timestamp")
	timestamp, err := strconv.ParseInt(timestampParam, 10, 64)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid timestamp: %v", timestampParam), http.StatusBadRequest)
	}
	employeeHistory, err := employees.AtPointInTime(ctx, h.Db, uint(id), time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("ID[%d]: %w", timestamp, err)
	}
	return web.Respond(ctx, w, employeeHistory, http.StatusOK)
}

// BetweenDates returns historical data for a given employee between dates.
func (h Handlers) BetweenDates(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	startTimestampParam := web.Param(r, "startTimestamp")
	startTimestamp, err := strconv.ParseInt(startTimestampParam, 10, 64)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid start timestamp: %v", startTimestampParam), http.StatusBadRequest)
	}
	endTimestampParam := web.Param(r, "endTimestamp")
	endTimestamp, err := strconv.ParseInt(endTimestampParam, 10, 64)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid end timestamp: %v", endTimestampParam), http.StatusBadRequest)
	}
	employeeHistory, err := employees.BetweenDates(ctx, h.Db, uint(id), time.Unix(startTimestamp, 0).Format("2006-01-02 15:04:05"), time.Unix(endTimestamp, 0).Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("ID[%d]: %w", startTimestamp, err)
	}
	return web.Respond(ctx, w, employeeHistory, http.StatusOK)
}
