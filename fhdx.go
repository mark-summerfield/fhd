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
		_, err = tx.CreateBucketIfNotExists(savesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", savesBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(saveInfoBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				saveInfoBucket, err)
		}
		return nil
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

func (me *Fhd) monitor(filenames ...string) ([]string, error) {
	missing := make([]string, 0)
	err := me.db.Update(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		var err error
		for _, filename := range filenames {
			filename = me.relativePath(filename)
			if !gong.FileExists(filename) {
				missing = append(missing, filename)
				continue // ignore nonexistent files
			}
			rawFilename := []byte(filename)
			var stateVal StateVal
			rawOldStateVal := states.Get(rawFilename)
			if rawOldStateVal != nil {
				stateVal = unmarshalStateVal(rawOldStateVal)
				stateVal.Monitored = true
			} else {
				stateVal = newStateVal(InvalidSID, true, binKind)
			}
			if ierr := states.Put(rawFilename,
				stateVal.marshal()); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
	return missing, err
}

func (me *Fhd) monitored(monitored bool) ([]*StateItem, error) {
	stateItems := make([]*StateItem, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		cursor := states.Cursor()
		rawFilename, rawStateVal := cursor.First()
		for ; rawFilename != nil; rawFilename,
			rawStateVal = cursor.Next() {
			stateVal := unmarshalStateVal(rawStateVal)
			if stateVal.Monitored == monitored {
				stateItems = append(stateItems,
					newState(string(rawFilename), stateVal))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stateItems, nil
}

func (me *Fhd) save(comment string, missing []string) (SaveResult, error) {
	monitored, err := me.Monitored()
	if err != nil {
		return newInvalidSaveResult(), err
	}
	var saveResult SaveResult
	err = me.db.Update(func(tx *bolt.Tx) error {
		var err error
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		ignores := me.getIgnores(tx)
		if ignores == nil {
			return fmt.Errorf("failed to find %q", configIgnore)
		}
		saveResult, err = me.nextSid(tx, comment)
		if err != nil {
			return err // saveResult is invalid on err
		}
		saveResult.MissingFiles = missing
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return errors.New("missing saves")
		}
		sid := saveResult.Sid
		save, err := saves.CreateBucket(sid.marshal())
		if err != nil {
			return fmt.Errorf("failed to save metadata for #%d", sid)
		}
		count := 0
		for _, stateItem := range monitored {
			if gong.FileExists(stateItem.Filename) { // Save
				saved, ierr := me.maybeSaveOne(tx, saves, save, sid,
					stateItem.Filename, stateItem.LastSid)
				if ierr != nil {
					err = errors.Join(err, ierr)
				}
				if saved {
					count++
				}
			} else { // Unmonitor
				saveResult.MissingFiles = append(saveResult.MissingFiles,
					stateItem.Filename)
				ierr := me.unmonitor(states, ignores, stateItem.Filename)
				if ierr != nil {
					err = errors.Join(err, ierr)
				}
			}
		}
		if err == nil && count > 0 {
			err = me.saveInfoItem(tx, saveResult.SaveInfoItem)
		}
		return err
	})
	return saveResult, err
}

func (me *Fhd) unmonitor(states, ignores *bolt.Bucket,
	filename string) error {
	rawFilename := []byte(me.relativePath(filename))
	rawOldStateVal := states.Get(rawFilename)
	if rawOldStateVal == nil { // Not Monitored so add to ignores
		return ignores.Put(rawFilename, emptyValue)
	} else {
		stateVal := unmarshalStateVal(rawOldStateVal)
		stateVal.Monitored = false
		return states.Put(rawFilename, stateVal.marshal())
	}
}

func (me *Fhd) getIgnores(tx *bolt.Tx) *bolt.Bucket {
	config := tx.Bucket(configBucket)
	if config == nil {
		return nil
	}
	return config.Bucket(configIgnore)
}

func (me *Fhd) nextSid(tx *bolt.Tx, comment string) (SaveResult, error) {
	var sid SID
	saveInfo := tx.Bucket(saveInfoBucket)
	if saveInfo == nil {
		return newInvalidSaveResult(), fmt.Errorf("failed to find %q",
			saveInfoBucket)
	}
	cursor := saveInfo.Cursor()
	rawSid, _ := cursor.Last()
	if rawSid == nil {
		sid = 1 // start at 1
	} else {
		sid = unmarshalSid(rawSid) + 1
	}
	return newSaveResult(sid, time.Now(), comment), nil
}

func (me *Fhd) saveInfoItem(tx *bolt.Tx, saveInfoItem SaveInfoItem) error {
	saveInfo := tx.Bucket(saveInfoBucket)
	if saveInfo == nil {
		return fmt.Errorf("failed to find %q", saveInfoBucket)
	}
	rawSaveInfoVal, err := saveInfoItem.SaveInfoVal.marshal()
	if err == nil {
		err = saveInfo.Put(saveInfoItem.Sid.marshal(), rawSaveInfoVal)
	}
	return err
}

// If the new file's SHA256 != prev SHA256 (or there is no prev) we save the
// file _and_ update the states with the SID for fast access to the file's
// most recent save.
func (me *Fhd) maybeSaveOne(tx *bolt.Tx, saves, save *bolt.Bucket, sid SID,
	filename string, prevSid SID) (bool, error) {
	var sha shA256
	raw, rawFlate, rawLzw, err := getRaws(filename, &sha)
	if err != nil {
		return false, err
	}
	if me.sameAsPrev(saves, sid, filename, prevSid, &sha) {
		return false, nil // No need to save if same as before.
	}
	compression := compressionForSizes(len(raw), len(rawFlate), len(rawLzw))
	saveVal := newSaveVal(sha, compression)
	switch compression {
	case noCompression:
		saveVal.Blob = raw
	case flateCompression:
		saveVal.Blob = rawFlate
	case lzwCompression:
		saveVal.Blob = rawLzw
	}
	rawFilename := []byte(filename)
	if err = save.Put(rawFilename, saveVal.marshal()); err != nil {
		return true, err
	}
	states := tx.Bucket(statesBucket)
	if states == nil {
		return true, errors.New("missing states")
	}
	stateVal := newStateVal(sid, true, fileKindForRaw(raw))
	return true, states.Put(rawFilename, stateVal.marshal())
}

func (me *Fhd) sameAsPrev(saves *bolt.Bucket, newSid SID, filename string,
	prevSid SID, newSha *shA256) bool {
	if prevSid == InvalidSID {
		return false
	}
	saveVal := me.getSaveVal(saves, filename, prevSid)
	if saveVal == nil {
		return false // There is no previous saveVal for this filename.
	}
	return saveVal.Sha == *newSha
}

func (me *Fhd) getSaveVal(saves *bolt.Bucket, filename string,
	sid SID) *saveVal {
	save := saves.Bucket(sid.marshal())
	if save == nil {
		return nil
	}
	rawSaveVal := save.Get([]byte(filename))
	if rawSaveVal == nil {
		return nil
	}
	return unmarshalSaveVal(rawSaveVal)
}

func (me *Fhd) relativePath(filename string) string {
	relPath, err := filepath.Rel(filepath.Dir(me.db.Path()), filename)
	if err != nil {
		return filepath.Clean(filename)
	}
	return relPath
}
