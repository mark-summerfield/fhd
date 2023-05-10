// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"net/http"
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

func (me *Fhd) monitor(filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		var err error
		for _, filename := range filenames {
			rawFilename := []byte(me.relativePath(filename))
			var stateInfo StateInfo
			rawOldStateInfo := states.Get(rawFilename)
			if rawOldStateInfo != nil {
				stateInfo = UnmarshalStateInfo(rawOldStateInfo)
				stateInfo.Monitored = true
			} else {
				stateInfo = newStateInfo(true, InvalidSID, "")
			}
			if ierr := states.Put(rawFilename,
				stateInfo.Marshal()); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
}

func (me *Fhd) monitored(monitored bool) ([]*StateData, error) {
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

func (me *Fhd) unmonitor(states, ignores *bolt.Bucket,
	filename string) error {
	rawFilename := []byte(me.relativePath(filename))
	rawOldStateInfo := states.Get(rawFilename)
	if rawOldStateInfo == nil { // Not Monitored so add to ignores
		return ignores.Put(rawFilename, emptyValue)
	} else {
		stateInfo := UnmarshalStateInfo(rawOldStateInfo)
		stateInfo.Monitored = false
		return states.Put(rawFilename, stateInfo.Marshal())
	}
}

func (me *Fhd) getIgnores(tx *bolt.Tx) *bolt.Bucket {
	config := tx.Bucket(configBucket)
	if config == nil {
		return nil
	}
	return config.Bucket(configIgnore)
}

func (me *Fhd) newSid(tx *bolt.Tx, comment string) (SaveInfo, error) {
	var sid SID
	saves := tx.Bucket(savesBucket)
	if saves == nil {
		return newInvalidSaveInfo(), fmt.Errorf("failed to find %q",
			savesBucket)
	}
	u, _ := saves.NextSequence()
	sid = SID(u)
	return newSaveInfo(sid, time.Now(), comment), nil
}

// If the new file's SHA256 != prev SHA256 (or there is no prev) we save the
// file _and_ update the states with the SID for fast access to the file's
// most recent save.
func (me *Fhd) maybeSaveOne(tx *bolt.Tx, saves, save *bolt.Bucket, sid SID,
	filename string, prevSid SID) (bool, error) {
	var sha SHA256
	raw, rawFlate, rawLzw, err := getRaws(filename, &sha)
	if err != nil {
		return false, err
	}
	if me.sameAsPrev(saves, sid, filename, prevSid, &sha) {
		return false, nil // No need to save if same as before.
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
	rawFilename := []byte(filename)
	if err = save.Put(rawFilename, entry.Marshal()); err != nil {
		return true, err
	}
	states := tx.Bucket(statesBucket)
	if states == nil {
		return true, errors.New("missing states")
	}
	stateInfo := newStateInfo(true, sid, http.DetectContentType(raw))
	return true, states.Put(rawFilename, stateInfo.Marshal())
}

func (me *Fhd) saveMetadata(save *bolt.Bucket, saveInfo *SaveInfo) error {
	rawWhen, err := saveInfo.When.MarshalBinary()
	if err != nil {
		return err
	}
	if err = save.Put(savesWhen, rawWhen); err != nil {
		return err
	}
	if err = save.Put(savesComment,
		[]byte(saveInfo.Comment)); err != nil {
		return err
	}
	return nil
}

func (me *Fhd) sameAsPrev(saves *bolt.Bucket, newSid SID, filename string,
	prevSid SID, newSha *SHA256) bool {
	if prevSid == InvalidSID {
		return false
	}
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

func (me *Fhd) relativePath(filename string) string {
	relPath, err := filepath.Rel(filepath.Dir(me.db.Path()), filename)
	if err != nil {
		return filepath.Clean(filename)
	}
	return relPath
}
