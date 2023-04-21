// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package employees

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db/employees/models"
)

var originalreadEmployeeHistory = readEmployeeHistory

func TestGetAllHistory(t *testing.T) {
	testCases := []struct {
		name                    string
		mockClosure             func() *sql.DB
		readEmployeeHistoryMock func(rows *sql.Rows, dest ...any) error
		expectedOutput          []models.EmployeeHistory
		expectedError           error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME ALL
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs(uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			expectedOutput: []models.EmployeeHistory{
				{
					Employee: models.Employee{
						Id:         1,
						FirstName:  "John",
						LastName:   "melo",
						Salary:     1234.56,
						Department: "it",
						Email:      "email@email.com",
					},
					RowStart: time.Time{},
					RowEnd:   time.Time{},
				},
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME ALL
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs(uint(1)).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error during scan",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME ALL
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs(uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			readEmployeeHistoryMock: func(rows *sql.Rows, dest ...any) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleReadEmployeeHistory(tc.readEmployeeHistoryMock)
			db := tc.mockClosure()
			output, err := GetAllHistory(context.TODO(), db, uint(1))
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

func TestAtPointInTime(t *testing.T) {
	testCases := []struct {
		name                    string
		mockClosure             func() *sql.DB
		readEmployeeHistoryMock func(rows *sql.Rows, dest ...any) error
		expectedOutput          []models.EmployeeHistory
		expectedError           error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				AS OF TIMESTAMP ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			expectedOutput: []models.EmployeeHistory{
				{
					Employee: models.Employee{
						Id:         1,
						FirstName:  "John",
						LastName:   "melo",
						Salary:     1234.56,
						Department: "it",
						Email:      "email@email.com",
					},
					RowStart: time.Time{},
					RowEnd:   time.Time{},
				},
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				AS OF TIMESTAMP ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", uint(1)).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error during scan",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				AS OF TIMESTAMP ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			readEmployeeHistoryMock: func(rows *sql.Rows, dest ...any) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleReadEmployeeHistory(tc.readEmployeeHistoryMock)
			db := tc.mockClosure()
			output, err := AtPointInTime(context.TODO(), db, uint(1), "2006-01-02 15:04:05")
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

func TestBetweenDates(t *testing.T) {
	testCases := []struct {
		name                    string
		mockClosure             func() *sql.DB
		readEmployeeHistoryMock func(rows *sql.Rows, dest ...any) error
		expectedOutput          []models.EmployeeHistory
		expectedError           error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				FROM ? TO ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", "2006-01-02 15:04:05", uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			expectedOutput: []models.EmployeeHistory{
				{
					Employee: models.Employee{
						Id:         1,
						FirstName:  "John",
						LastName:   "melo",
						Salary:     1234.56,
						Department: "it",
						Email:      "email@email.com",
					},
					RowStart: time.Time{},
					RowEnd:   time.Time{},
				},
			},
		},
		{
			name: "error",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				FROM ? TO ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", "2006-01-02 15:04:05", uint(1)).
					WillReturnError(errors.New("random error"))
				return db
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error during scan",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `
				SELECT id, first_name, last_name, salary, department, email, row_start, row_end
				FROM employees
				FOR SYSTEM_TIME
				FROM ? TO ?
				WHERE id = ?
				`
				mock.ExpectQuery(regexp.QuoteMeta(q)).
					WithArgs("2006-01-02 15:04:05", "2006-01-02 15:04:05", uint(1)).
					WillReturnRows(sqlmock.NewRows(
						[]string{"id", "first_name", "last_name", "salary", "department", "email", "row_start", "row_end"}).
						AddRow(1, "John", "melo", "1234.56", "it", "email@email.com", time.Time{}, time.Time{}))
				return db
			},
			readEmployeeHistoryMock: func(rows *sql.Rows, dest ...any) error {
				return errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleReadEmployeeHistory(tc.readEmployeeHistoryMock)
			db := tc.mockClosure()
			output, err := BetweenDates(context.TODO(), db, uint(1), "2006-01-02 15:04:05", "2006-01-02 15:04:05")
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

func handleReadEmployeeHistory(mocked func(rows *sql.Rows, dest ...any) error) {
	if mocked != nil {
		readEmployeeHistory = mocked
	} else {
		readEmployeeHistory = originalreadEmployeeHistory
	}
}
