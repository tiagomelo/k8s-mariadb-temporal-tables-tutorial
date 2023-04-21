// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package employees

import (
	"context"
	"database/sql"

	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"
)

// For ease of unit testing.
var readEmployeeHistory = func(rows *sql.Rows, dest ...any) error {
	return rows.Scan(dest...)
}

// GetAllHistory returns the complete history of a given employee.
func GetAllHistory(ctx context.Context, db *sql.DB, id uint) ([]models.EmployeeHistory, error) {
	q := `
	SELECT id, first_name, last_name, salary, department, email, row_start, row_end
	FROM employees
	FOR SYSTEM_TIME ALL
	WHERE id = ?
	`
	employeeHistory := make([]models.EmployeeHistory, 0)
	rows, err := db.QueryContext(ctx, q, id)
	if err != nil {
		return employeeHistory, err
	}
	defer rows.Close()
	for rows.Next() {
		var employeeHist models.EmployeeHistory
		if err = readEmployeeHistory(rows,
			&employeeHist.Id,
			&employeeHist.FirstName,
			&employeeHist.LastName,
			&employeeHist.Salary,
			&employeeHist.Department,
			&employeeHist.Email,
			&employeeHist.RowStart,
			&employeeHist.RowEnd,
		); err != nil {
			return nil, err
		}
		employeeHistory = append(employeeHistory, employeeHist)
	}
	return employeeHistory, nil
}

// AtPointInTime returns the complete history of a given employee in a given
// point in time.
func AtPointInTime(ctx context.Context, db *sql.DB, id uint, timestamp string) ([]models.EmployeeHistory, error) {
	q := `
	SELECT id, first_name, last_name, salary, department, email, row_start, row_end
	FROM employees
	FOR SYSTEM_TIME
	AS OF TIMESTAMP ?
	WHERE id = ?
	`
	employeeHistory := make([]models.EmployeeHistory, 0)
	rows, err := db.QueryContext(ctx, q, timestamp, id)
	if err != nil {
		return employeeHistory, err
	}
	defer rows.Close()
	for rows.Next() {
		var employeeHist models.EmployeeHistory
		if err = readEmployeeHistory(rows,
			&employeeHist.Id,
			&employeeHist.FirstName,
			&employeeHist.LastName,
			&employeeHist.Salary,
			&employeeHist.Department,
			&employeeHist.Email,
			&employeeHist.RowStart,
			&employeeHist.RowEnd,
		); err != nil {
			return nil, err
		}
		employeeHistory = append(employeeHistory, employeeHist)
	}
	return employeeHistory, nil
}

// BetweenDates returns the complete history of a given employee in a given period.
func BetweenDates(ctx context.Context, db *sql.DB, id uint, startTimestamp, endTimeStamp string) ([]models.EmployeeHistory, error) {
	q := `
	SELECT id, first_name, last_name, salary, department, email, row_start, row_end
	FROM employees
	FOR SYSTEM_TIME
	FROM ? TO ?
	WHERE id = ?
	`
	employeeHistory := make([]models.EmployeeHistory, 0)
	rows, err := db.QueryContext(ctx, q, startTimestamp, endTimeStamp, id)
	if err != nil {
		return employeeHistory, err
	}
	defer rows.Close()
	for rows.Next() {
		var employeeHist models.EmployeeHistory
		if err = readEmployeeHistory(rows,
			&employeeHist.Id,
			&employeeHist.FirstName,
			&employeeHist.LastName,
			&employeeHist.Salary,
			&employeeHist.Department,
			&employeeHist.Email,
			&employeeHist.RowStart,
			&employeeHist.RowEnd,
		); err != nil {
			return nil, err
		}
		employeeHistory = append(employeeHistory, employeeHist)
	}
	return employeeHistory, nil
}
