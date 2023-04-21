// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package employees

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	mariaDb "github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"
)

var originalReadEmployees = readEmployees

func TestGetById(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		expectedOutput *models.Employee
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				return db
			},
			expectedOutput: &models.Employee{
				Id:         1,
				FirstName:  "John",
				LastName:   "Doe",
				Salary:     1234.56,
				Department: "it",
				Email:      "email@email.com",
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnError(mariaDb.ErrDBNotFound)
				return db
			},
			expectedError: errors.New("not found"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := GetById(context.TODO(), db, uint(1))
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		scanRowsMock   func(rows *sql.Rows, dest ...any) error
		expectedOutput []models.Employee
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				return db
			},
			expectedOutput: []models.Employee{
				{
					Id:         1,
					FirstName:  "John",
					LastName:   "Doe",
					Salary:     1234.56,
					Department: "it",
					Email:      "email@email.com",
				},
			},
		},
		{
			name: "error during scan",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				return db
			},
			scanRowsMock: func(rows *sql.Rows, dest ...any) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleScanRows(tc.scanRowsMock)
			db := tc.mockClosure()
			output, err := GetAll(context.TODO(), db)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		expectedOutput *models.Employee
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				INSERT INTO
					employees(first_name, last_name, salary, department, email)
				VALUES
					(?, ?, ?, ?, ?)
				RETURNING
					id, first_name, last_name, salary, department, email
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("John", "Doe", 1234.56, "it", "email@email.com").
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				return db
			},
			expectedOutput: &models.Employee{
				Id:         1,
				FirstName:  "John",
				LastName:   "Doe",
				Salary:     1234.56,
				Department: "it",
				Email:      "email@email.com",
			},
		},
		{
			name: "error - duplicated entry",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				INSERT INTO
					employees(first_name, last_name, salary, department, email)
				VALUES
					(?, ?, ?, ?, ?)
				RETURNING
					id, first_name, last_name, salary, department, email
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("John", "Doe", 1234.56, "it", "email@email.com").
					WillReturnError(&mysql.MySQLError{
						Number: mariaDb.UniqueViolation,
					})
				return db
			},
			expectedError: errors.New("duplicated entry"),
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				INSERT INTO
					employees(first_name, last_name, salary, department, email)
				VALUES
					(?, ?, ?, ?, ?)
				RETURNING
					id, first_name, last_name, salary, department, email
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs("John", "Doe", 1234.56, "it", "email@email.com").
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := Create(context.TODO(), db, &models.NewEmployee{
				FirstName:  "John",
				LastName:   "Doe",
				Salary:     1234.56,
				Department: "it",
				Email:      "email@email.com",
			})
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	testCases := []struct {
		name           string
		mockClosure    func() *sql.DB
		expectedOutput *models.Employee
		expectedError  error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				uq := `
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
				mock.ExpectExec(regexp.QuoteMeta(uq)).WithArgs("John", "Doe", 1234.56, "it", "email@email.com", 1).WillReturnResult(driver.ResultNoRows)
				return db
			},
			expectedOutput: &models.Employee{
				Id:         1,
				FirstName:  "John",
				LastName:   "Doe",
				Salary:     1234.56,
				Department: "it",
				Email:      "email@email.com",
			},
		},
		{
			name: "error - not found",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnError(mariaDb.ErrDBNotFound)
				return db
			},
			expectedError: errors.New("not found"),
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email
				FROM employees
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email"}).
						AddRow(1, "John", "Doe", "1234.56", "it", "email@email.com"))
				uq := `
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
				mock.ExpectExec(regexp.QuoteMeta(uq)).
					WithArgs("John", "Doe", 1234.56, "it", "email@email.com", 1).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			output, err := Update(context.TODO(), db, uint(1), &models.UpdateEmployee{
				FirstName:  stringPointer("John"),
				LastName:   stringPointer("Doe"),
				Salary:     float64Pointer(1234.56),
				Department: stringPointer("it"),
				Email:      stringPointer("email@email.com"),
			})
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Equal(t, tc.expectedOutput, output)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name          string
		mockClosure   func() *sql.DB
		expectedError error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				DELETE FROM
					employees
				WHERE id = ?
				`
				mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnResult(driver.ResultNoRows)
				return db
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				DELETE FROM
					employees
				WHERE id = ?
				`
				mock.ExpectExec(regexp.QuoteMeta(q)).WithArgs(1).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			err := Delete(context.TODO(), db, uint(1))
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.Nil(t, err)
			}
		})
	}
}

func stringPointer(s string) *string {
	return &s
}

func float64Pointer(f float64) *float64 {
	return &f
}

func handleScanRows(mocked func(rows *sql.Rows, dest ...any) error) {
	if mocked != nil {
		readEmployees = mocked
	} else {
		readEmployees = originalReadEmployees
	}
}
