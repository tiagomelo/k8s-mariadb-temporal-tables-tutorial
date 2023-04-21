// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

var (
	originalCurrentDateProvider = currentDate
	originalRandomIntProvider   = randomInt
)

func TestConnectToPostgres(t *testing.T) {
	testCases := []struct {
		name          string
		mockSqlOpen   func(driverName string, dataSourceName string) (*sql.DB, error)
		expectedError error
	}{
		{
			name: "happy path",
			mockSqlOpen: func(driverName, dataSourceName string) (*sql.DB, error) {
				db, _, _ := sqlmock.New()
				return db, nil
			},
		},
		{
			name: "error",
			mockSqlOpen: func(driverName, dataSourceName string) (*sql.DB, error) {
				return nil, errors.New("random error")
			},
			expectedError: errors.New("random error"),
		},
		{
			name: "error pinging",
			mockSqlOpen: func(driverName, dataSourceName string) (*sql.DB, error) {
				db, mock, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
				mock.ExpectPing().WillReturnError(errors.New("random error"))
				return db, nil
			},
			expectedError: errors.New("random error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlOpen = tc.mockSqlOpen
			db, err := ConnectToMariaDb("user", "pass", "host", "port", "schema")
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
				require.NotNil(t, db)
			}
		})
	}
}

func TestAdvanceMariaDbTimestamp(t *testing.T) {
	testCases := []struct {
		name                string
		currentDateProvider func() time.Time
		randomIntProvider   func() int
		mockClosure         func() *sql.DB
		expectedError       error
	}{
		{
			name: "happy path",
			currentDateProvider: func() time.Time {
				t, _ := time.Parse(time.DateOnly, "2023-01-01")
				return t
			},
			randomIntProvider: func() int {
				return 1
			},
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `SET @@timestamp = UNIX_TIMESTAMP('2024-01-01')`
				mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnResult(driver.ResultNoRows)
				return db
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handleCurrentDateProvider(tc.currentDateProvider)
			handleRandomIntProvider(tc.randomIntProvider)
			db := tc.mockClosure()
			err := AdvanceMariaDbTimestamp(context.TODO(), db)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
			}
		})
	}
}

func TestSetDefaultMariaDbTimestamp(t *testing.T) {
	testCases := []struct {
		name          string
		mockClosure   func() *sql.DB
		expectedError error
	}{
		{
			name: "happy path",
			mockClosure: func() *sql.DB {
				db, mock, _ := sqlmock.New()
				q := `SET @@timestamp = default`
				mock.ExpectExec(regexp.QuoteMeta(q)).WillReturnResult(driver.ResultNoRows)
				return db
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := tc.mockClosure()
			err := SetDefaultMariaDbTimestamp(context.TODO(), db)
			if err != nil {
				if tc.expectedError == nil {
					t.Fatalf(`expected no error, got "%v"`, err)
				}
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				if tc.expectedError != nil {
					t.Fatalf(`expected error "%v", got nil`, tc.expectedError)
				}
			}
		})
	}
}

func handleCurrentDateProvider(mocked func() time.Time) {
	if mocked != nil {
		currentDate = mocked
	} else {
		currentDate = originalCurrentDateProvider
	}
}

func handleRandomIntProvider(mocked func() int) {
	if mocked != nil {
		randomInt = mocked
	} else {
		randomInt = originalRandomIntProvider
	}
}
