// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// setState sets the state of every given file the the given state except as
// folows.
// If state is Ignored: if a file's current state is Monitored, its state
// will be set to Unmonitored, and if its current state is Renamed, its
// state won't change.
// Can only go from Monitored to Unmonitored, not Ignored.
func (me *Fhd) setState(state StateKind, filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		buck := tx.Bucket(stateBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", stateBucket)
		}
		var err error
		for _, filename := range filenames {
			key := []byte(filename)
			newState := state
			oldState := StateKind(buck.Get(key))
			if oldState != nil {
				if newState.Equal(Unmonitored) && oldState.Equal(Ignored) {
					continue // Ignored is implicitly Unmonitored
				} else if newState.Equal(Ignored) {
					if oldState.Equal(Renamed) {
						continue // Can't go from Renamed to Ignored
					}
					if oldState.Equal(Monitored) {
						newState = Unmonitored
					}
				}
			}
			if ierr := buck.Put(key, newState); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
}

func (me *Fhd) stateOf(state StateKind) ([]string, error) {
	monitored := make([]string, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		buck := tx.Bucket(stateBucket)
		if buck == nil {
			return fmt.Errorf("failed to find %q", stateBucket)
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

func (me *Fhd) nextSid(comment string) SidInfo {
	sid := 0 // TODO make valid SID
	return newSidInfo(sid, time.Now(), comment)
}

func (me *Fhd) saveOne(sid int, filename string) error {
	return fmt.Errorf("saveOne unimplemented %d %q", sid, filename) // TODO
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
}
