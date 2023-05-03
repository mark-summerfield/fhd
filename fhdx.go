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
// will be set to Unmonitored.
// Can only go from Monitored to Unmonitored, not Ignored.
func (me *Fhd) setState(state StateKind, filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		var err error
		for _, filename := range filenames {
			key := []byte(filename)
			newState := state
			oldState := StateKind(states.Get(key))
			if oldState != nil {
				if newState.Equal(Unmonitored) && oldState.Equal(Ignored) {
					continue // Ignored is implicitly Unmonitored
				} else if newState.Equal(Ignored) &&
					oldState.Equal(Monitored) {
					newState = Unmonitored
				}
			}
			if ierr := states.Put(key, newState); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
}

func (me *Fhd) hasState(state StateKind) ([]string, error) {
	monitored := make([]string, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		cursor := states.Cursor()
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

func (me *Fhd) nextSid(comment string) (SidInfo, error) {
	var sid uint64
	err := me.db.Update(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		sid, _ = saves.NextSequence()
		return nil
	})
	if err != nil {
		return newInvalidSidInfo(), err
	}
	return newSidInfo(sid, time.Now(), comment), nil
}

func (me *Fhd) maybeSaveOne(sid uint64, filename string) error {
	return fmt.Errorf("maybeSaveOne unimplemented %d %q", sid, filename) // TODO
	/*
		data = read filename's content
		create 3 goroutines to compute data's
			- sha256
			- flate
			- lzw
		find previous sha256
		if previous sha256 == sha256:
			return nil # don't save duplicate
		flag := flagForSizes(len(raw), len(flate), len(lzw))
		switch {
			case flag == Raw: # new content, no benefit from compression
				blob = new content
			case flag == Flate:
				blob = flate
			case flag == Lzw
				blob = lzw
		}
		return nil
	*/
}
