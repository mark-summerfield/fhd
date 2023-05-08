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
		err := makeConfig(tx)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(statesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				statesBucket, err)
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

func makeConfig(tx *bolt.Tx) error {
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
	ignores, err := config.CreateBucketIfNotExists(configIgnore)
	if err != nil {
		return fmt.Errorf("failed to create bucket %q: %s",
			configIgnore, err)
	}
	for _, filename := range defaultIgnores {
		if ierr := ignores.Put([]byte(filename),
			emptyValue); ierr != nil {
			err = errors.Join(err, ierr)
		}
	}
	return err
}

func (me *Fhd) setMonitored(filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		var err error
		for _, filename := range filenames {
			key := []byte(me.relativePath(filename))
			var sid SID
			rawOldStateInfo := states.Get(key)
			if rawOldStateInfo != nil {
				oldStateInfo := UnmarshalStateInfo(rawOldStateInfo)
				sid = oldStateInfo.Sid
			}
			stateInfo := newStateInfo(true, sid)
			if ierr := states.Put(key, stateInfo.Marshal()); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
}

// setUnmonitored sets the state of every given file to Unmonitored if it is
// being monitored and preserves its SID. For any file that isn't already
// Monitored, adds it to the config/ignore list.
func (me *Fhd) setUnmonitored(filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		ignores := me.getIgnores(tx)
		if ignores == nil {
			return fmt.Errorf("failed to find %q", configIgnore)
		}
		var err error
		for _, filename := range filenames {
			key := []byte(me.relativePath(filename))
			rawOldStateInfo := states.Get(key)
			if rawOldStateInfo == nil { // Not Monitored so add to ignores
				if ierr := ignores.Put(key, emptyValue); ierr != nil {
					err = errors.Join(err, ierr)
				}
			} else {
				oldStateInfo := UnmarshalStateInfo(rawOldStateInfo)
				stateInfo := newStateInfo(false, oldStateInfo.Sid)
				if ierr := states.Put(key,
					stateInfo.Marshal()); ierr != nil {
					err = errors.Join(err, ierr)
				}
			}
		}
		return err
	})
}

func (me *Fhd) areMonitored(monitored bool) ([]*StateData, error) {
	stateData := make([]*StateData, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		cursor := states.Cursor()
		rawFilename, rawStateInfo := cursor.First()
		for ; rawFilename != nil; rawFilename,
			rawStateInfo = cursor.Next() {
			stateInfo := UnmarshalStateInfo(rawStateInfo)
			if stateInfo.Monitored == monitored {
				stateData = append(stateData, newState(string(rawFilename),
					stateInfo))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stateData, nil
}

func (me *Fhd) setIgnored(filenames ...string) error {
	err := me.db.Update(func(tx *bolt.Tx) error {
		ignores := me.getIgnores(tx)
		var err error
		for _, filename := range filenames {
			if ierr := ignores.Put([]byte(filename),
				emptyValue); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
	return err
}

func (me *Fhd) areIgnored() ([]string, error) {
	ignored := make([]string, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		ignores := me.getIgnores(tx)
		if ignores == nil {
			return fmt.Errorf("failed to find %q", configIgnore)
		}
		cursor := ignores.Cursor()
		rawFilename, _ := cursor.First()
		for ; rawFilename != nil; rawFilename, _ = cursor.Next() {
			ignored = append(ignored, string(rawFilename))
		}
		return nil
	})
	return ignored, err
}

func (me *Fhd) getIgnores(tx *bolt.Tx) *bolt.Bucket {
	config := tx.Bucket(configBucket)
	if config == nil {
		return nil
	}
	return config.Bucket(configIgnore)
}

func (me *Fhd) newSid(comment string) (SaveInfo, error) {
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
		return newInvalidSaveInfo(), err
	}
	return newSaveInfo(sid, time.Now(), comment), nil
}

// If the new file's SHA256 != prev SHA256 (or there is no prev) we save the
// file _and_ update the states with the SID for fast access to the file's
// most recent save.
func (me *Fhd) maybeSaveOne(tx *bolt.Tx, sid SID, filename string,
	prevSid SID) error {
	saves := tx.Bucket(savesBucket)
	if saves == nil {
		return errors.New("missing saves")
	}
	var sha SHA256
	raw, rawFlate, rawLzw, err := getRaws(filename, &sha)
	if err != nil {
		return err
	}
	if me.sameAsPrev(saves, sid, filename, prevSid, &sha) {
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
	if err = saves.Put(sid.Marshal(), entry.Marshal()); err == nil {
		return err
	}
	states := tx.Bucket(statesBucket)
	if states == nil {
		return errors.New("missing states")
	}
	stateInfo := newStateInfo(true, sid)
	return states.Put([]byte(filename), stateInfo.Marshal())
}

func (me *Fhd) sameAsPrev(saves *bolt.Bucket, newSid SID, filename string,
	prevSid SID, newSha *SHA256) bool {
	entry := me.getEntry(saves, filename, prevSid)
	if entry == nil {
		return false // There is no previous entry for this filename.
	}
	return entry.Sha == *newSha
}

func (me *Fhd) getEntry(saves *bolt.Bucket, filename string,
	sid SID) *Entry {
	save := saves.Bucket(sid.Marshal())
	if save == nil {
		return nil
	}
	rawEntry := save.Get([]byte(filename))
	if rawEntry == nil {
		return nil
	}
	return UnmarshalEntry(rawEntry)
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
