// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package employees

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	mariaDb "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/validate"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web"
	v1Web "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/web/v1"
)

type Handlers struct {
	Db *sql.DB
}

// handleGetEmployeeByIdError handles errors when getting an
// employee by its id.
func handleGetEmployeeByIdError(err error, id uint) error {
	if errors.Is(err, mariaDb.ErrDBNotFound) {
		return v1Web.NewRequestError(err, http.StatusNotFound)
	}
	return fmt.Errorf("ID[%d]: %w", id, err)
}

// handleCreateEmployeeError handles errors when creating an
// employee.
func handleCreateEmployeeError(err error) error {
	if errors.Is(err, mariaDb.ErrDBDuplicatedEntry) {
		return v1Web.NewRequestError(err, http.StatusConflict)
	}
	return fmt.Errorf("unable to create employee: %w", err)
}

// handleUpdateEmployeeByIdErr handles errors when updating an
// employee by its id.
func handleUpdateEmployeeByIdErr(err error, id uint) error {
	if errors.Is(err, mariaDb.ErrDBNotFound) {
		return v1Web.NewRequestError(err, http.StatusNotFound)
	}
	return fmt.Errorf("ID[%d]: %w", id, err)
}

// GetById returns a current employee with given id.
func (h Handlers) GetById(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	employee, err := employees.GetById(ctx, h.Db, uint(id))
	if err != nil {
		return handleGetEmployeeByIdError(err, uint(id))
	}
	return web.Respond(ctx, w, employee, http.StatusOK)
}

// GetAll returns all current employees.
func (h *Handlers) GetAll(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	employees, err := employees.GetAll(ctx, h.Db)
	if err != nil {
		return fmt.Errorf("unable to query employees: %w", err)
	}
	return web.Respond(ctx, w, employees, http.StatusOK)
}

// Create creates an employee.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	var newEmployee models.NewEmployee
	if err := json.NewDecoder(r.Body).Decode(&newEmployee); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}
	if err := validate.Check(newEmployee); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}
	employee, err := employees.Create(ctx, h.Db, &newEmployee)
	if err != nil {
		return handleCreateEmployeeError(err)
	}
	return web.Respond(ctx, w, employee, http.StatusCreated)
}

// Update updates a current employee.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	var updateEmployee models.UpdateEmployee
	if err := json.NewDecoder(r.Body).Decode(&updateEmployee); err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}
	updatedEmployee, err := employees.Update(ctx, h.Db, uint(id), &updateEmployee)
	if err != nil {
		return handleUpdateEmployeeByIdErr(err, uint(id))
	}
	return web.Respond(ctx, w, updatedEmployee, http.StatusOK)
}

// Delete deletes a current employee.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	idParam := web.Param(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid id: %v", idParam), http.StatusBadRequest)
	}
	if err = employees.Delete(ctx, h.Db, uint(id)); err != nil {
		return fmt.Errorf("ID[%d]: %w", id, err)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
