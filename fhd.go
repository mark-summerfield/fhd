// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"database/sql"

	"github.com/mark-summerfield/gong"
	_ "github.com/mattn/go-sqlite3"
)

type Fhd struct {
	db *sql.DB
}

// New opens (and creates if necessary) the given .fhd file ready for use.
func New(filename string) (*Fhd, error) {
	filename = gong.AbsPath(filename)
	create := !gong.FileExists(filename)
	db, err := sql.Open(dbDriverName, filename)
	if err != nil {
		return nil, err
	}
	err = setPragmas(db)
	if err != nil {
		return nil, err
	}
	if create {
		err = makeTables(db)
		if err != nil {
			return nil, err
		}
	}
	return &Fhd{db: db}, nil
}

// Close closes the underlying database.
func (me *Fhd) Close() error {
	return me.db.Close()
}
