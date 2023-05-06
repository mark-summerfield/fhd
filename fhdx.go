// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/mark-summerfield/gong"
	bolt "go.etcd.io/bbolt"
)

func newDb(filename string) (*bolt.DB, error) {
	db, err := bolt.Open(filename, gong.ModeUserRW, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		config, err := tx.CreateBucketIfNotExists(configBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				configBucket, err)
		}
		if format := config.Get(configFormat); len(format) != 1 {
			if err = config.Put(configFormat,
				[]byte{fileFormat}); err != nil {
				return fmt.Errorf("failed to initialize %q file format: %s",
					configBucket, err)
			}
		}
		_, err = tx.CreateBucketIfNotExists(statesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				statesBucket, err)
		}
		_, err = tx.CreateBucketIfNotExists(renamedBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				renamedBucket, err)
		}
		saves, err := tx.CreateBucketIfNotExists(savesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", savesBucket,
				err)
		}
		return saves.SetSequence(0) // next i.e., first used, will be 1.
	})
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return nil, errors.Join(err, closeErr)
		}
		return nil, err
	}
	return db, nil
}

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
			key := []byte(me.relativePath(filename))
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

func (me *Fhd) haveState(state StateKind) ([]string, error) {
	filenames := make([]string, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		cursor := states.Cursor()
		rawFilename, rawState := cursor.First()
		for ; rawFilename != nil; rawFilename, rawState = cursor.Next() {
			if state.Equal(StateKind(rawState)) {
				filenames = append(filenames, string(rawFilename))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filenames, nil
}

func (me *Fhd) newSid(comment string) (SidInfo, error) {
	var sid SID
	err := me.db.Update(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		u, _ := saves.NextSequence()
		sid = SID(u)
		return nil
	})
	if err != nil {
		return newInvalidSidInfo(), err
	}
	return newSidInfo(sid, time.Now(), comment), nil
}

func (me *Fhd) maybeSaveOne(saves *bolt.Bucket, sid SID,
	filename string) error {
	var sha SHA256
	raw, rawFlate, rawLzw, err := getRaws(filename, &sha)
	if err != nil {
		return err
	}
	if me.sameAsPrev(saves, sid, filename, &sha) {
		return nil // No need to save if same as before.
	}
	flag := flagForSizes(len(raw), len(rawFlate), len(rawLzw))
	entry := newEntry(sha, flag)
	switch flag {
	case Raw:
		entry.Blob = raw
	case Flate:
		entry.Blob = rawFlate
	case Lzw:
		entry.Blob = rawLzw
	}
	return saves.Put(MarshalSid(sid), entry.Marshal())
}

func (me *Fhd) sameAsPrev(saves *bolt.Bucket, newSid SID, filename string,
	newSha *SHA256) bool {
	entry := me.findLatestEntry(saves, filename)
	if entry == nil {
		return false // There is no previous entry for this filename.
	}
	return entry.Sha == *newSha
}

func (me *Fhd) findLatestEntry(saves *bolt.Bucket, filename string) *Entry {
	cursor := saves.Cursor()
	rawFilename, rawEntry := cursor.Last()
	for ; rawFilename != nil; rawFilename, rawEntry = cursor.Prev() {
		if string(rawFilename) == filename {
			return UnmarshalEntry(rawEntry)
		}
	}
	return nil
}

func (me *Fhd) relativePath(filename string) string {
	relPath, err := filepath.Rel(filepath.Dir(me.db.Path()), filename)
	if err != nil {
		return filepath.Clean(filename)
	}
	return relPath
}
