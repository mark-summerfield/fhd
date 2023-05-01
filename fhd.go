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
	db, err := bolt.Open(filename, ModeUserRW, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		buc, err := tx.CreateBucketIfNotExists(ConfigBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				ConfigBucket, err)
		}
		if format := buc.Get(ConfigFormat); len(format) != 1 {
			if err = buc.Put(ConfigFormat, []byte{FileFormat}); err != nil {
				return fmt.Errorf("failed to initialize %q file format: %s",
					ConfigBucket, err)
			}
		}
		_, err = tx.CreateBucketIfNotExists(StateBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", StateBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(SavesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", SavesBucket,
				err)
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

// Format returns the underlying database's file format number.
func (me *Fhd) FileFormat() (int, error) {
	var fileFormat byte
	err := me.db.View(func(tx *bolt.Tx) error {
		format := tx.Bucket(ConfigBucket).Get(ConfigFormat)
		if len(format) == 1 {
			fileFormat = format[0]
		}
		return nil
	})
	if err != nil {
		return int(FileFormat), err
	} else if fileFormat == 0 {
		return int(FileFormat), nil
	}
	return int(fileFormat), nil
}

// State returns the state of every known file.
func (me *Fhd) State() ([]*StateData, error) {
	stateData := make([]*StateData, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket(StateBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", StateBucket)
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
// Note that if state is Ignored and a file is being monitored, that file's
// state will be set to Unmonitored.
func (me *Fhd) SetState(state StateKind, filenames []string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		buck := tx.Bucket(StateBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", StateBucket)
		}
		var err error
		for _, filename := range filenames {
			key := []byte(filename)
			newState := state
			oldState := buck.Get(key)
			if oldState != nil && state.Equal(Ignored) {
				newState = Unmonitored
			}
			if ierr := buck.Put(key, newState); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
}

// Monitored returns the list of every monitored file.
func (me *Fhd) Monitored() ([]string, error) {
	return me.stateOf(Monitored)
}

// Save saves a snapshot of every monitored file.
func (me *Fhd) Save() error {
	monitored, err := me.Monitored()
	if err != nil {
		return err
	}
	for _, filename := range monitored {
		if ierr := me.saveOne(filename); ierr != nil {
			err = errors.Join(err, ierr)
		}
	}
	return err
}

func (me *Fhd) stateOf(state StateKind) ([]string, error) {
	monitored := make([]string, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket(StateBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", StateBucket)
		}
		cursor := buck.Cursor()
		rawFilename, rawState := cursor.First()
		for ; rawFilename != nil; rawFilename, rawState = cursor.Next() {
			if state.Equal(StateKind(rawState)) {
				monitored = append(monitored, string(rawFilename))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return monitored, nil
}

func (me *Fhd) saveOne(filename string) error {
	fmt.Println("saveOne", filename)
	/*
		new data = read filename's content
		find filename in saves (should be in last save, i.e., most recent sid, foundSid)
		create 2 or 3 goroutines
			- compute sha256 for new data -> ([]byte, Raw)
			- gzip data -> ([]byte, Gz)
			- if found, patch (diff, old data) -> ([]byte, Patch)
		flag := flagForSizes(len(raw), len(gz), len(patch))
		switch {
			case new sha256 == old sha256: # unchanged data
				blob = sha256 → empty
				flag = InOld
				oldSid = foundSid
				oldFilename = empty
			case flag == Raw: # new content
				blob = new content
				flag = Raw
				sha256 = new sha256
				oldSid = 0
				oldFilename = empty
			case flag == Gz:
				blob = gzipped
				sha256 = new sha256 # to check ungzip
				flag = Gz
				oldSid = 0
				oldFilename = empty
			case flag == Patch
				blob = patch
				sha256 = new sha256 # to check old blob + patch
				flag = Patch
				oldSid = foundSid
				oldFilename = empty
			}

	*/
	return nil
}

// Extract
// func (me *Fhd) Extract(filename string) (string, error) {
//	// must a/c for patches
//	// returns new filename e.g. filename#1.ext and err
// }

func (me *Fhd) findSid(filename string) (int, error) {
	var sid int
	err := me.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket(SavesBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", SavesBucket)
		}
		// TODO iterate key (sid) from last back to first using cursor
		// check value bucket for matching filename & if found set sid &
		// break
		return nil
	})
	return sid, err
}
