// Copyright (c) 2022 Tiago Melo. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.
package dbtest

import (
	"database/sql"
	"fmt"
)

// SetCurrentTimeStamp overrides the default MariaDB's timestamp.
func SetCurrentTimeStamp(db *sql.DB, timestamp string) error {
	q := fmt.Sprintf("SET @@timestamp = UNIX_TIMESTAMP('%s')", timestamp)
	_, err := db.Exec(q)
	return err
}

// SetDefaultTimeStamp sets back the default MariaDB's timestamp.
func SetDefaultTimeStamp(db *sql.DB) error {
	q := "SET @@timestamp = default"
	_, err := db.Exec(q)
	return err
}
