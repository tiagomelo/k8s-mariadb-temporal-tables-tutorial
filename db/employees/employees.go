// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package employees

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	mariaDb "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"
)

// For ease of unit testing.
var (
	readEmployee = func(row *sql.Row, dest ...any) error {
		return row.Scan(dest...)
	}
	readEmployees = func(rows *sql.Rows, dest ...any) error {
		return rows.Scan(dest...)
	}
)

// GetById returns a current employee with given id.
func GetById(ctx context.Context, db *sql.DB, id uint) (*models.Employee, error) {
	q := `
	SELECT id, first_name, last_name, salary, department, email
	FROM employees
	WHERE id = ?
	`
	var employee models.Employee
	row := db.QueryRowContext(ctx, q, id)
	if err := readEmployee(row,
		&employee.Id,
		&employee.FirstName,
		&employee.LastName,
		&employee.Salary,
		&employee.Department,
		&employee.Email,
	); err != nil {
		return nil, mariaDb.ErrDBNotFound
	}
	return &employee, nil
}

// GetAll returns all current employees.
func GetAll(ctx context.Context, db *sql.DB) ([]models.Employee, error) {
	q := `
	SELECT id, first_name, last_name, salary, department, email
	FROM employees
	`
	employees := make([]models.Employee, 0)
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return employees, err
	}
	defer rows.Close()
	for rows.Next() {
		var employee models.Employee
		if err = readEmployees(rows,
			&employee.Id,
			&employee.FirstName,
			&employee.LastName,
			&employee.Salary,
			&employee.Department,
			&employee.Email,
		); err != nil {
			return employees, err
		}
		employees = append(employees, employee)
	}
	return employees, nil
}

// Create creates an employee.
func Create(ctx context.Context, db *sql.DB,
	newEmployee *models.NewEmployee) (*models.Employee, error) {
	q := `
	INSERT INTO
		employees(first_name, last_name, salary, department, email)
	VALUES
		(?, ?, ?, ?, ?)
	RETURNING
		id, first_name, last_name, salary, department, email
	`
	var employee models.Employee
	row := db.QueryRowContext(ctx, q, newEmployee.FirstName,
		newEmployee.LastName, newEmployee.Salary, newEmployee.Department, newEmployee.Email)
	if err := readEmployee(row,
		&employee.Id,
		&employee.FirstName,
		&employee.LastName,
		&employee.Salary,
		&employee.Department,
		&employee.Email,
	); err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == mariaDb.UniqueViolation {
			return nil, mariaDb.ErrDBDuplicatedEntry
		}
		return nil, err
	}
	return &employee, nil
}

// handleFirstNameChanges handles changes to FirstName.
func handleFirstNameChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	if updateEmployee.FirstNameIsFulfilled() {
		dbEmployee.FirstName = *updateEmployee.FirstName
	}
}

// handleLastNameChanges handles changes to LastName.
func handleLastNameChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	if updateEmployee.LastNameIsFulfilled() {
		dbEmployee.LastName = *updateEmployee.LastName
	}
}

// handleDepartmentChanges handles changes to Department.
func handleDepartmentChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	if updateEmployee.DepartmentIsFulfilled() {
		dbEmployee.Department = *updateEmployee.Department
	}
}

// handleEmailChanges handles changes to Email.
func handleEmailChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	if updateEmployee.EmailIsFulfilled() {
		dbEmployee.Email = *updateEmployee.Email
	}
}

// handleSalaryChanges handles changes to Salary.
func handleSalaryChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	if updateEmployee.SalaryIsFulfilled() {
		dbEmployee.Salary = *updateEmployee.Salary
	}
}

// handleEmployeeChanges updates the changed properties.
func handleEmployeeChanges(updateEmployee *models.UpdateEmployee, dbEmployee *models.Employee) {
	handleFirstNameChanges(updateEmployee, dbEmployee)
	handleLastNameChanges(updateEmployee, dbEmployee)
	handleDepartmentChanges(updateEmployee, dbEmployee)
	handleEmailChanges(updateEmployee, dbEmployee)
	handleSalaryChanges(updateEmployee, dbEmployee)
}

// Update updates an employee.
func Update(ctx context.Context, db *sql.DB, employeeId uint,
	updateEmployee *models.UpdateEmployee) (*models.Employee, error) {
	q := `
	UPDATE employees
	SET 
		first_name = ?,
		last_name = ?,
		salary = ?,
		department = ?,
		email = ?
	WHERE
		id = ?
	`
	dbEmployee, err := GetById(ctx, db, employeeId)
	if err != nil {
		return nil, err
	}
	handleEmployeeChanges(updateEmployee, dbEmployee)
	_, err = db.ExecContext(ctx, q, dbEmployee.FirstName,
		dbEmployee.LastName, dbEmployee.Salary, dbEmployee.Department,
		dbEmployee.Email, dbEmployee.Id)
	return dbEmployee, err
}

// Delete deletes an employee.
func Delete(ctx context.Context, db *sql.DB, id uint) error {
	q := `
	DELETE FROM
		employees
	WHERE id = ?
	`
	_, err := db.ExecContext(ctx, q, id)
	return err
}
