// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	randomInt = func() int {
		min := 1
		max := 5
		return rand.Intn(max-min+1) + min
	}
	currentDate = func() time.Time {
		return time.Now().UTC()
	}
)

// Set of error codes from MariaDB.
const (
	UniqueViolation = uint16(1062)
	NotFoundErr     = uint16(1054)
)

// Set of error variables for CRUD operations.
var (
	// sqlOpen eases unit testing.
	sqlOpen              = sql.Open
	ErrDBNotFound        = errors.New("not found")
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
)

// ConnectToMariaDb establishes a connection to Postgres.
func ConnectToMariaDb(user, pass, host, port, schema string) (*sql.DB, error) {
	dsnString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, schema)
	db, err := sqlOpen("mysql", dsnString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// AdvanceMariaDbTimestamp advances timestamp in database. It is used mainly for demonstration
// purposes.
func AdvanceMariaDbTimestamp(ctx context.Context, db *sql.DB) error {
	ts := currentDate().AddDate(randomInt(), 0, 0)
	q := fmt.Sprintf(`SET @@timestamp = UNIX_TIMESTAMP('%s')`, ts.Format(time.DateOnly))
	_, err := db.ExecContext(ctx, q)
	return err
}

// SetDefaultMariaDbTimestamp sets the database timestamp back to default, which is,
// the current date.
func SetDefaultMariaDbTimestamp(ctx context.Context, db *sql.DB) error {
	q := `SET @@timestamp = default`
	_, err := db.ExecContext(ctx, q)
	return err
}
