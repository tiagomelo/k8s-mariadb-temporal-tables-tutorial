//go:build integration

// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package test

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/db"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/handlers"
	"github.com/tiagomelo/k8s-mariadb-temporal-tables-tutorial/test/dbtest"
	"go.uber.org/zap"
)

var (
	api             http.Handler
	testDb          *sql.DB
	firstTimeStamp  = stringPointer("2022-01-01")
	secondTimeStamp = stringPointer("2022-05-01 11:00:00")
)

func tearDown(db *sql.DB, log *zap.SugaredLogger) {
	if err := dbtest.SetDefaultTimeStamp(testDb); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := db.Exec("ALTER TABLE employees DROP SYSTEM VERSIONING"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := db.Exec("TRUNCATE TABLE employees"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := db.Exec("ALTER TABLE employees ADD SYSTEM VERSIONING"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := db.Close(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := log.Sync(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	mariaDbUser := os.Getenv("MARIADB_USER")
	mariaDbPassword := os.Getenv("MARIADB_PASSWORD")
	mariaDbHost := os.Getenv("MARIADB_HOST_NAME")
	mariaDbPort := os.Getenv("MARIADB_PORT")
	mariaDbDatabase := os.Getenv("MARIADB_DATABASE")
	var err error
	testDb, err = db.ConnectToMariaDb(
		mariaDbUser, mariaDbPassword,
		mariaDbHost, mariaDbPort, mariaDbDatabase,
	)
	if err != nil {
		fmt.Println("error when connecting to the test database:", err)
		os.Exit(1)
	}
	shutdown := make(chan os.Signal, 1)
	log := zap.NewNop().Sugar()
	api = handlers.NewAPIMux(handlers.APIMuxConfig{
		Shutdown: shutdown,
		Db:       testDb,
		Log:      log,
	})
	exitVal := m.Run()
	tearDown(testDb, log)
	os.Exit(exitVal)
}

func TestCreateEmployee(t *testing.T) {
	testCases := []struct {
		name                 string
		input                []byte
		timestamp            *string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name:      "happy path",
			timestamp: firstTimeStamp,
			input: []byte(`{
				"first_name": "John",
				"last_name": "Doe",
				"department": "IT",
				"salary": 1234.56,
				"email": "email@email.com"
			}`),
			expectedResponse: `
			{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"department": "IT",
				"salary": 1234.56,
				"email": "email@email.com"
			}
			`,
			expectedResponseCode: http.StatusCreated,
		},
		{
			name:  "invalid json payload",
			input: []byte(`invalid`),
			expectedResponse: `
			{
				"error": "invalid character 'i' looking for beginning of value"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:  "missing required fields",
			input: []byte(`{}`),
			expectedResponse: `
			{
				"error": "data validation error",
				"fields": {
					"department": "department is a required field",
					"first_name": "first_name is a required field",
					"last_name": "last_name is a required field",
					"salary": "salary is a required field",
					"email": "email is a required field"
				}
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "duplicated entry",
			input: []byte(`{
				"first_name": "John",
				"last_name": "Doe",
				"department": "IT",
				"salary": 1234.56,
				"email": "email@email.com"
			}`),
			expectedResponse: `
			{
				"error": "duplicated entry"
			}
			`,
			expectedResponseCode: http.StatusConflict,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.timestamp != nil {
				if err := dbtest.SetCurrentTimeStamp(testDb, *tc.timestamp); err != nil {
					t.Fatalf(`error setting timestamp %s`, *tc.timestamp)
				}
			}
			req := httptest.NewRequest(http.MethodPost, "/v1/employee", bytes.NewBuffer(tc.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())

		})
	}
}

func TestGetAllEmployees(t *testing.T) {
	testCases := []struct {
		name                 string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name: "happy path",
			expectedResponse: `
			[
				{
					"department":"IT", 
					"first_name":"John", 
					"id":1, 
					"last_name":"Doe", 
					"salary":1234.56,
					"email": "email@email.com"
				}
			]`,
			expectedResponseCode: http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/employees", nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestGetEmployeeById(t *testing.T) {
	testCases := []struct {
		name                 string
		id                   string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name: "happy path",
			id:   "1",
			expectedResponse: `
			{
				"department":"IT", 
				"first_name":"John", 
				"id":1, 
				"last_name":"Doe", 
				"salary":1234.56,
				"email": "email@email.com"
			}
			`,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "invalid id param",
			id:   "1a",
			expectedResponse: `
			{
				"error": "invalid id: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "not found",
			id:   "666",
			expectedResponse: `
			{
				"error": "not found"
			}
			`,
			expectedResponseCode: http.StatusNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/employee/%s", tc.id), nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	testCases := []struct {
		name                 string
		id                   string
		timestamp            *string
		input                []byte
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name:      "happy path",
			id:        "1",
			timestamp: secondTimeStamp,
			input: []byte(`{
				"salary": 3000.56
			}`),
			expectedResponse: `
			{
				"id": 1,
				"first_name": "John",
				"last_name": "Doe",
				"department": "IT",
				"salary": 3000.56,
				"email": "email@email.com"
			}
			`,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "invalid id param",
			id:   "1a",
			expectedResponse: `
			{
				"error": "invalid id: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:  "invalid json payload",
			input: []byte(`invalid`),
			id:    "1",
			expectedResponse: `
			{
				"error": "invalid character 'i' looking for beginning of value"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "not found",
			input: []byte(`{
				"last_name": "Deer"
			}`),
			id: "666",
			expectedResponse: `
			{
				"error": "not found"
			}
			`,
			expectedResponseCode: http.StatusNotFound,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.timestamp != nil {
				if err := dbtest.SetCurrentTimeStamp(testDb, *tc.timestamp); err != nil {
					t.Fatalf(`error setting timestamp %s`, *tc.timestamp)
				}
			}
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/employee/%s", tc.id), bytes.NewBuffer(tc.input))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestGetAllEmployeeHistory(t *testing.T) {
	testCases := []struct {
		name                 string
		id                   string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name: "happy path",
			id:   "1",
			expectedResponse: `
			[
				{
					"department": "IT",
					"first_name": "John",
					"id": 1,
					"last_name": "Doe",
					"row_end": "2022-05-01T11:00:00Z",
					"row_start": "2022-01-01T00:00:00Z",
					"salary": 1234.56,
					"email": "email@email.com"
				},
				{
					"department": "IT",
					"first_name": "John",
					"id": 1,
					"last_name": "Doe",
					"row_end": "2038-01-19T03:14:07.999999Z",
					"row_start": "2022-05-01T11:00:00Z",
					"salary": 3000.56,
					"email": "email@email.com"
				}
			]
			`,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "invalid id param",
			id:   "1a",
			expectedResponse: `
			{
				"error": "invalid id: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/employee/%s/history/all", tc.id), nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestGetEmployeHistoryAtPointInTime(t *testing.T) {
	testCases := []struct {
		name                 string
		id                   string
		timestamp            string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name:      "happy path",
			id:        "1",
			timestamp: "1643725560",
			expectedResponse: `
			[
				{
					"department": "IT",
					"first_name": "John",
					"id": 1,
					"last_name": "Doe",
					"row_end": "2022-05-01T11:00:00Z",
					"row_start": "2022-01-01T00:00:00Z",
					"salary": 1234.56,
					"email": "email@email.com"
				}
			]
			`,
			expectedResponseCode: http.StatusOK,
		},
		{
			name:      "invalid id param",
			id:        "1a",
			timestamp: "1675434360",
			expectedResponse: `
			{
				"error": "invalid id: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:      "invalid timestamp param",
			id:        "1",
			timestamp: "1a",
			expectedResponse: `
			{
				"error": "invalid timestamp: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/employee/%s/history/%s", tc.id, tc.timestamp), nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestGetEmployeHistoryBetweenTwoDates(t *testing.T) {
	testCases := []struct {
		name                 string
		id                   string
		startTimestamp       string
		endTimestamp         string
		expectedResponse     string
		expectedResponseCode int
	}{
		{
			name:           "happy path",
			id:             "1",
			startTimestamp: "1643725560",
			endTimestamp:   "1672588815",
			expectedResponse: `
			[
				{
					"department": "IT",
					"first_name": "John",
					"id": 1,
					"last_name": "Doe",
					"row_end": "2022-05-01T11:00:00Z",
					"row_start": "2022-01-01T00:00:00Z",
					"salary": 1234.56,
					"email": "email@email.com"
				},
				{
					"department": "IT",
					"first_name": "John",
					"id": 1,
					"last_name": "Doe",
					"row_end": "2038-01-19T03:14:07.999999Z",
					"row_start": "2022-05-01T11:00:00Z",
					"salary": 3000.56,
					"email": "email@email.com"
				}
			]
			`,
			expectedResponseCode: http.StatusOK,
		},
		{
			name:           "invalid id param",
			id:             "1a",
			startTimestamp: "1643725560",
			endTimestamp:   "1672588815",
			expectedResponse: `
			{
				"error": "invalid id: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:           "invalid start timestamp param",
			id:             "1",
			startTimestamp: "1a",
			endTimestamp:   "1672588815",
			expectedResponse: `
			{
				"error": "invalid start timestamp: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:           "invalid end timestamp param",
			id:             "1",
			startTimestamp: "1643725560",
			endTimestamp:   "1a",
			expectedResponse: `
			{
				"error": "invalid end timestamp: 1a"
			}
			`,
			expectedResponseCode: http.StatusBadRequest,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/employee/%s/history/%s/%s", tc.id, tc.startTimestamp, tc.endTimestamp), nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
			require.JSONEq(t, tc.expectedResponse, w.Body.String())
		})
	}
}

func TestAdvanceTimestamp(t *testing.T) {
	testCases := []struct {
		name                 string
		expectedResponseCode int
	}{
		{
			name:                 "happy path",
			expectedResponseCode: http.StatusNoContent,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/db/timestamp/advance", nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
		})
	}
}

func TestSetDefaultTimestamp(t *testing.T) {
	testCases := []struct {
		name                 string
		expectedResponseCode int
	}{
		{
			name:                 "happy path",
			expectedResponseCode: http.StatusNoContent,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/db/timestamp/default", nil)
			w := httptest.NewRecorder()
			api.ServeHTTP(w, req)
			require.Equal(t, tc.expectedResponseCode, w.Code)
		})
	}
}

func stringPointer(s string) *string {
	return &s
}
