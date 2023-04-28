// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Fhd struct {
	db *bolt.DB
}

// New opens (and creates if necessary) the given .fhd file ready for use.
func New(filename string) (*Fhd, error) {
	db, err := bolt.Open(filename, ModeOwnerRW, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(StateBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", StateBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(SavesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", SavesBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(RenamedBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				RenamedBucket, err)
		}
		return nil
	})
	if err != nil {
		closeEerr := db.Close()
		if closeEerr != nil {
			return nil, errors.Join(err, closeEerr)
		}
		return nil, err
	}
	return &Fhd{db: db}, nil
}

// Close closes the underlying database.
func (me *Fhd) Close() error {
	return me.db.Close()
}

// Filename returns the underlying database's filename.
func (me *Fhd) Filename() string {
	return me.db.Path()
}

// State returns the state of every known file.
func (me *Fhd) State() ([]*StateData, error) {
	stateData := make([]*StateData, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket(StateBucket)
		if buck == nil {
			return errors.New("failed to find StateBucket")
		}
		cursor := buck.Cursor()
		rawFilename, rawState := cursor.First()
		for ; rawFilename != nil; rawFilename, rawState = cursor.Next() {
			stateData = append(stateData, newStateFromRaw(rawFilename,
				rawState))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stateData, nil
}

// SetState sets the state of every given file the the given state.
func (me *Fhd) SetState(state StateKind, filenames []string) error {
	// if state == Ignored but filename is in fhd then for that filename set
	// state to be Unmonitored ??? Or leave that as higher-level logic for
	// callers ???

	return nil // TODO
}
