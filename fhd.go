// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"io"

	bolt "go.etcd.io/bbolt"
)

type Fhd struct {
	db *bolt.DB
}

// New opens (and creates if necessary) the given .fhd file ready for use.
func New(filename string) (*Fhd, error) {
	db, err := newDb(filename)
	if err != nil {
		return nil, err
	}
	return &Fhd{db: db}, nil
}

// Close closes the underlying database.
func (me *Fhd) Close() error {
	return me.db.Close()
}

// See also Dump.
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

// States returns the state of every known file.
// See also Monitored, Unmonitored, and Ignored.
func (me *Fhd) States() ([]*StateData, error) {
	stateData := make([]*StateData, 0)
	err := me.db.View(func(tx *bolt.Tx) error {
		states := tx.Bucket(statesBucket)
		if states == nil {
			return fmt.Errorf("failed to find %q", statesBucket)
		}
		cursor := states.Cursor()
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

// Monitored returns the list of every monitored file.
// See also State.
func (me *Fhd) Monitored() ([]string, error) {
	return me.haveState(Monitored)
}

// Unmonitored returns the list of every unmonitored file.
// See also State.
func (me *Fhd) Unmonitored() ([]string, error) {
	return me.haveState(Unmonitored)
}

// Ignored returns the list of every ignored file.
// See also State.
func (me *Fhd) Ignored() ([]string, error) {
	return me.haveState(Ignored)
}

// Monitor sets the given files to be monitored _and_ does an initial Save.
// Returns the new Save ID (SID).
// See also MonitorWithComment, Unmonitor and Ignore.
func (me *Fhd) Monitor(filenames ...string) (SidInfo,
	error) {
	err := me.setState(Monitored, filenames...)
	if err != nil {
		return newInvalidSidInfo(), err
	}
	return me.Save("")
}

// MonitorWithComment sets the given files to be monitored _and_ does an
// initial Save. Returns the new Save ID (SID).
// See also Monitor, Unmonitor and Ignore.
func (me *Fhd) MonitorWithComment(comment string,
	filenames ...string) (SidInfo, error) {
	err := me.setState(Monitored, filenames...)
	if err != nil {
		return newInvalidSidInfo(), err
	}
	return me.Save(comment)
}

// Unmonitor sets the given files to be unmonitored. Any Ignored files stay
// Ignored.
// See also Monitor and Ignore.
func (me *Fhd) Unmonitor(filenames ...string) error {
	return me.setState(Unmonitored, filenames...)
}

// Ignore sets the given files to be ignored. Any Monitored files become
// Unmonitored rather than Ignored.
// See also Monitor and Unmonitor.
func (me *Fhd) Ignore(filenames ...string) error {
	return me.setState(Ignored, filenames...)
}

// Save saves a snapshot of every monitored file that's changed, and returns
// the corresponding Save ID (SID).
func (me *Fhd) Save(comment string) (SidInfo, error) {
	monitored, err := me.Monitored()
	if err != nil {
		return newInvalidSidInfo(), err
	}
	sidInfo, err := me.newSid(comment)
	if err != nil {
		return newInvalidSidInfo(), err
	}
	sid := sidInfo.Sid()
	err = me.db.Update(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return errors.New("missing saves")
		}
		var err error
		for _, filename := range monitored {
			if ierr := me.maybeSaveOne(saves, sid, filename); ierr != nil {
				err = errors.Join(err, ierr)
			}
		}
		return err
	})
	return sidInfo, err
}

// Returns the most recent Save ID (SID).
func (me *Fhd) Sid() (SidInfo, error) {
	var sidInfo SidInfo
	err := me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		cursor := saves.Cursor()
		sid, _ := cursor.Last()
		save := saves.Bucket(sid)
		if save != nil {
			rawWhen := save.Get(savesWhen)
			when, err := timeForRaw(rawWhen)
			if err != nil {
				return err
			}
			var comment string
			rawComment := save.Get(savesComment)
			if rawComment != nil {
				comment = string(rawComment)
			}
			u, err := btou(sid)
			if err != nil {
				return err
			}
			sidInfo = newSidInfo(u, when, comment)
		}
		return nil
	})
	if err != nil {
		return newInvalidSidInfo(), err
	}
	return sidInfo, nil
}

// Returns all the Save IDs (SIDs).
func (me *Fhd) Sids() ([]SidInfo, error) {
	sidInfos := make([]SidInfo, 0)
	return sidInfos, errors.New("Sids unimplemented") // TODO
}

// Returns the most recent SID for the given filename.
func (me *Fhd) SidForFilename(filename string) (SidInfo, error) {
	var sidInfo SidInfo
	err := me.db.View(func(tx *bolt.Tx) error {
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			return fmt.Errorf("failed to find %q", savesBucket)
		}
		// TODO iterate key (sid) from last back to first using cursor
		// check value bucket for matching filename & if found set sid &
		// break
		return nil
	})
	return sidInfo, err
}

// Returns the all the SIDs for the given filename.
func (me *Fhd) SidsForFilename(filename string) ([]SidInfo, error) {
	sidInfos := make([]SidInfo, 0)
	return sidInfos, errors.New("SidsForFilename unimplemented") // TODO
}

// Writes the content of the given filename from the most recent Save to the
// given writer.
func (me *Fhd) Extract(filename string, writer io.Writer) error {
	sidInfo, err := me.SidForFilename(filename)
	if err != nil {
		return err
	}
	return me.ExtractForSid(sidInfo.Sid(), filename, writer)
}

// Writes the content of the given filename from the specified Save
// (identified by its SID) to the given writer.
func (me *Fhd) ExtractForSid(sid uint64, filename string,
	writer io.Writer) error {
	return errors.New("ExtractForSid unimplemented") // TODO
}

// Rename renames oldFilename to newFilename.
func (me *Fhd) Rename(oldFilename, newFilename string) error {
	return errors.New("Rename unimplemented") // TODO
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
	return errors.New("Delete unimplemented") // TODO
}

// Purges deletes every save of the given file and sets the file's state is
// set to Ignored.
func (me *Fhd) Purge(filename string) error {
	return errors.New("Purge unimplemented") // TODO
}
