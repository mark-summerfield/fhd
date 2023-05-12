// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"database/sql"
	"errors"
)

func setPragmas(db *sql.DB) error {
	return execStatements(db, dbPragmas)
}

func makeTables(db *sql.DB) error {
	return execStatements(db, dbTables)
}

func execStatements(db *sql.DB, statements []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, statement := range statements {
		_, err = db.Exec(statement)
		if err != nil {
			ierr := tx.Rollback()
			if ierr != nil {
				return errors.Join(err, ierr)
			}
			return err
		}
	}
	return tx.Commit()
}
