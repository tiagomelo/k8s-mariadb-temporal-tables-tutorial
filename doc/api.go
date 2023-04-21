// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package doc

import "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"

// swagger:route GET /v1/employees employees GetAll
// Get all current employees.
// ---
// responses:
//		200: getAllCurrentEmployeesResponse

// swagger:response getAllCurrentEmployeesResponse
type GetAllCurrentEmployeesResponseWrapper struct {
	// in:body
	Body []models.Employee
}

// swagger:route GET /v1/employee/{id} employee GetById
// Get a current employee by its id.
// ---
// responses:
//
//	200: employee
//	400: description: invalid id
//	404: description: employee not found
//
// swagger:parameters GetById
type GetEmployeeByIdParamsWrapper struct {
	// in:path
	Id int
}

// swagger:response employee
type EmployeeResponseWrapper struct {
	// in:body
	Body models.Employee
}

// swagger:route POST /v1/employee employee Create
// Create an employee.
// ---
// responses:
//
//	200: employee
//
// swagger:parameters Create
type PostEmployeeParamsWrapper struct {
	// in:body
	Body models.NewEmployee
}

// swagger:route PUT /v1/employee/{id} employee Update
// Updates a current employee.
// ---
// responses:
//
//	200: employee
//	400: description: invalid id
//	404: description: employee not found
//
// swagger:parameters Update
type PutEmployeeParamsWrapper struct {
	// in:path
	Id int
	// in:body
	Body models.UpdateEmployee
}

// swagger:route DELETE /v1/employee/{id} employee Delete
// Deletes a current employee.
// ---
// responses:
//
//	204: description: no content
//	400: description: invalid id
//
// swagger:parameters Delete
type DeleteEmployeeParamsWrapper struct {
	// in:path
	Id int
}

// swagger:route GET /v1/employee/{id}/history/all history GetAllEmployeeHistoryById
// Get all historical data about an employee with a given id.
// ---
// responses:
//
//	200: employeeHistory
//	400: description: invalid id
//
// swagger:parameters GetAllEmployeeHistoryById
type GetAllEmployeeHistoryByIdParamsWrapper struct {
	// in:path
	Id int
}

// swagger:route GET /v1/employee/{id}/history/{timestamp} history GetAllEmployeeHistoryAtPointInTime
// Get historical data about an employee with a given id at a given point in time.
// ---
// responses:
//
//	200: employeeHistory
//	400: description: invalid id
//	400: description: invalid timestamp
//
// swagger:parameters GetAllEmployeeHistoryAtPointInTime
type GetAllEmployeeHistoryAtPointInTimeParamsWrapper struct {
	// in:path
	Timestamp int
}

// swagger:route GET /v1/employee/{id}/history/{startTimestamp}/{endTimestamp} history GetAllEmployeeHistoryBetweenDates
// Get historical data about an employee with a given id between dates.
// ---
// responses:
//
//	200: employeeHistory
//	400: description: invalid id
//	400: description: invalid start timestamp
//	400: description: invalid end timestamp
//
// swagger:parameters GetAllEmployeeHistoryBetweenDates
type GetAllEmployeeHistoryBetweenDatesParamsWrapper struct {
	// in:path
	StartTimestamp int
	// in:path
	EndTimestamp int
}

// swagger:response employeeHistory
type EmployeeHistoryResponseWrapper struct {
	// in:body
	Body []models.EmployeeHistory
}
