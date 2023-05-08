// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"compress/flate"
	"compress/lzw"
	"errors"
	"fmt"
	"io"

	"github.com/mark-summerfield/gong"
	bolt "go.etcd.io/bbolt"
)

type Fhd struct {
	db *bolt.DB
}

// New opens (and creates if necessary) the given .fhd file ready for use.
func New(filename string) (*Fhd, error) {
	db, err := newDb(gong.AbsPath(filename))
	if err != nil {
		return nil, err
	}
	return &Fhd{db: db}, nil
}

// Close closes the underlying database.
func (me *Fhd) Close() error {
	return me.db.Close()
}

func (me *Fhd) String() string {
	format, _ := me.FileFormat()
	return fmt.Sprintf("<Fhd filename=%q format=%d>", me.db.Path(), format)
}

// Filename returns the underlying database's filename.
func (me *Fhd) Filename() string {
	return me.db.Path()
}

// Format returns the underlying database's file format number.
func (me *Fhd) FileFormat() (int, error) {
	var fileformat byte
	err := me.db.View(func(tx *bolt.Tx) error {
		format := tx.Bucket(configBucket).Get(configFormat)
		if len(format) == 1 {
			fileformat = format[0]
		}
		return nil
	})
	if err != nil {
		return int(fileFormat), err
	} else if fileformat == 0 {
		return int(fileFormat), nil
	}
	return int(fileformat), nil
}

// States returns the monitoring state of every monitored and unmonitored
// file and the SID of the last save it was saved into.
func (me *Fhd) States() ([]*StateData, error) {
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
			stateData = append(stateData, newStateFromRaw(rawFilename,
				rawStateInfo))
		}
		return nil
	})
	return stateData, err
}

// Monitored returns the list of every monitored file.
func (me *Fhd) Monitored() ([]*StateData, error) {
	return me.monitored(true)
}

// Monitor adds the given files to be monitored _and_ does an initial Save.
// Returns the new Save ID (SID).
func (me *Fhd) Monitor(filenames ...string) (SaveInfo, error) {
	return me.MonitorWithComment("", filenames...)
}

// MonitorWithComment adds the given files to be monitored _and_ does an
// initial Save with the given comment. Returns the new Save ID (SID).
func (me *Fhd) MonitorWithComment(comment string,
	filenames ...string) (SaveInfo, error) {
	err := me.monitor(filenames...)
	if err != nil {
		return newInvalidSaveInfo(), err
	}
	return me.Save(comment)
}

// Unmonitored returns the list of every unmonitored file.
// These are files that have been monitored in the past but have been set to
// be unmonitored.
func (me *Fhd) Unmonitored() ([]*StateData, error) {
	return me.monitored(false)
}

// Unmonitor sets the state of every given file to unmonitored if it is
// being monitored and preserves its SID. For any file that isn't already
// monitored, adds it to the ignored list.
func (me *Fhd) Unmonitor(filenames ...string) error {
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

// Ignored returns the list of every ignored filename or glob.
func (me *Fhd) Ignored() ([]string, error) {
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

// Ignore adds the given files or globs to the ignored list.
func (me *Fhd) Ignore(filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
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
}

// Unignore deletes the given filenames or globs from the ignored list.
// But it never deletes "*.fhd".
func (me *Fhd) Unignore(filenames ...string) error {
	return me.db.Update(func(tx *bolt.Tx) error {
		ignores := me.getIgnores(tx)
		var err error
		for _, filename := range filenames {
			if filename != "*.fhd" {
				if ierr := ignores.Delete([]byte(filename)); ierr != nil {
					err = errors.Join(err, ierr)
				}
			}
		}
		return err
	})
}

// Save saves a snapshot of every monitored file that's changed and returns
// the corresponding SaveInfo with the new save ID (SID). If no files were
// changed and therefore none were saved, SaveInfo is invalid and err is
// nil.
func (me *Fhd) Save(comment string) (SaveInfo, error) {
	monitored, err := me.Monitored()
	if err != nil {
		return newInvalidSaveInfo(), err
	}
	var saveInfo SaveInfo
	err = me.db.Update(func(tx *bolt.Tx) error {
		var err error
		saveInfo, err = me.newSid(tx, comment)
		if err != nil {
			return err // saveInfo is invalid on err
		}
		sid := saveInfo.Sid
		for _, stateData := range monitored {
			if ierr := me.maybeSaveOne(tx, sid, stateData.Filename,
				stateData.Sid); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return errors.New("missing saves")
		}
		return me.saveMetadata(saves.Bucket(sid.Marshal()), &saveInfo)
	})
	return saveInfo, err
}

// SaveInfo returns the SaveInfo for the given SID or an invalid SaveInfo on
// error.
func (me *Fhd) SaveInfo(sid SID) SaveInfo {
	var saveInfo SaveInfo
	err := me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		save := saves.Bucket(sid.Marshal())
		if save != nil {
			rawWhen := save.Get(savesWhen)
			when, err := UnmarshalTime(rawWhen)
			if err != nil {
				return err
			}
			var comment string
			rawComment := save.Get(savesComment)
			if rawComment != nil {
				comment = string(rawComment)
			}
			saveInfo = newSaveInfo(sid, when, comment)
		}
		return nil
	})
	if err != nil {
		return newInvalidSaveInfo()
	}
	return saveInfo
}

// Returns the most recent Save ID (SID) or 0 on error.
func (me *Fhd) Sid() SID {
	var sid SID
	_ = me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves != nil {
			cursor := saves.Cursor()
			rawSid, _ := cursor.Last()
			sid = UnmarshalSid(rawSid)
		}
		return nil
	})
	return sid
}

// Returns all the Save IDs (SIDs) from most- to least-recent.
func (me *Fhd) Sids() ([]SID, error) {
	sids := make([]SID, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		cursor := saves.Cursor()
		rawSid, _ := cursor.Last()
		for ; rawSid != nil; rawSid, _ = cursor.Prev() {
			sids = append(sids, UnmarshalSid(rawSid))
		}
		return nil
	})
	return sids, err
}

// Returns the most recent SID for the given filename.
func (me *Fhd) SidForFilename(filename string) (SID, error) {
	rawFilename := []byte(me.relativePath(filename))
	var sid SID
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		rawStateInfo := states.Get(rawFilename)
		if rawStateInfo != nil {
			stateInfo := UnmarshalStateInfo(rawStateInfo)
			sid = stateInfo.Sid
		}
		return nil
	})
	return sid, err
}

// Returns all the SIDs for the given filename from most- to least-recent.
func (me *Fhd) SidsForFilename(filename string) ([]SID, error) {
	rawFilename := []byte(me.relativePath(filename))
	sids := make([]SID, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		cursor := saves.Cursor()
		rawSid, _ := cursor.Last()
		for ; rawSid != nil; rawSid, _ = cursor.Prev() {
			if save := saves.Bucket(rawFilename); save != nil {
				sids = append(sids, UnmarshalSid(rawSid))
			}
		}
		return nil
	})
	return sids, err
}

// Writes the content of the given filename from the most recent Save to the
// given writer.
func (me *Fhd) Extract(filename string, writer io.Writer) error {
	filename = me.relativePath(filename)
	sid, err := me.SidForFilename(filename)
	if err != nil {
		return err
	}
	return me.ExtractForSid(sid, filename, writer)
}

// Writes the content of the given filename from the specified Save
// (identified by its SID) to the given writer.
func (me *Fhd) ExtractForSid(sid SID, filename string,
	writer io.Writer) error {
	rawFilename := []byte(me.relativePath(filename))
	return me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		save := saves.Bucket(sid.Marshal())
		if save == nil {
			return fmt.Errorf("failed to find save %d", sid)
		}
		rawEntry := save.Get(rawFilename)
		if rawEntry == nil {
			return fmt.Errorf("failed to find file %s in save %d", filename,
				sid)
		}
		entry := UnmarshalEntry(rawEntry)
		var err error
		rawReader := bytes.NewReader(entry.Blob)
		switch entry.Flag {
		case Raw:
			_, err = io.Copy(writer, rawReader)
		case Flate:
			flateReader := flate.NewReader(rawReader)
			_, err = io.Copy(writer, flateReader)
		case Lzw:
			lzwReader := lzw.NewReader(rawReader, lzw.MSB, 0)
			_, err = io.Copy(writer, lzwReader)
		default:
			return fmt.Errorf("invalid flag %v", entry.Flag)
		}
		return err
	})
}

// Compact eliminates wasted space in the .fhd file.
func (me *Fhd) Compact() error {
	// temp := me.db.Path() + ".$$$"
	// me.db.CopyFile(temp)
	// move me.db.Path() to tempdir/tempname
	// rename temp dropping ".$$$"
	// delete tempdir/tempname
	return errors.New("Compact unimplemented") // TODO
}

// Delete deletes the given file in the given save.
// If this is the only occurrence of the file, the file's state is set to
// Ignored.
func (me *Fhd) Delete(sid int, filename string) error {
	//filename = me.relativePath(filename) // TODO
	return errors.New("Delete unimplemented") // TODO
}

// Purges deletes every save of the given file and sets the file's state is
// set to Ignored.
func (me *Fhd) Purge(filename string) error {
	//filename = me.relativePath(filename) // TODO
	return errors.New("Purge unimplemented") // TODO
}
