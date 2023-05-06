// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"fmt"
	"io"
	"time"

	"golang.org/x/exp/slices"

	bolt "go.etcd.io/bbolt"
)

type (
	WriteStr func(string)
	WriteRaw func([]byte)
)

// Dump writes data from the underlying database to the writer purely for
// debugging and testing.
func (me *Fhd) Dump(writer io.Writer) error {
	write := func(text string) { _, _ = writer.Write([]byte(text)) }
	writeRaw := func(raw []byte) { _, _ = writer.Write(raw) }
	return me.db.View(func(tx *bolt.Tx) error {
		dumpConfig(tx, write, writeRaw)
		dumpStates(tx, write, writeRaw)
		dumpRenamed(tx, write, writeRaw)
		return dumpSaves(tx, write, writeRaw)
	})
}

func dumpConfig(tx *bolt.Tx, write WriteStr, writeRaw WriteRaw) {
	config := tx.Bucket(configBucket)
	if config == nil {
		write("error: missing config\n")
	} else {
		format := config.Get(configFormat)
		write("config\n  format=")
		if len(format) == 0 {
			write("error (nil)")
		} else {
			write(fmt.Sprintf("%d", format[0]))
		}
		write("\n")
	}
}

func dumpStates(tx *bolt.Tx, write WriteStr, writeRaw WriteRaw) {
	states := tx.Bucket(statesBucket)
	if states == nil {
		write("error: missing states\n")
	} else {
		write("states:\n")
		cursor := states.Cursor()
		rawFilename, rawStateInfo := cursor.First()
		for ; rawFilename != nil; rawFilename,
			rawStateInfo = cursor.Next() {
			write("  ")
			writeRaw(rawFilename)
			stateInfo := UnmarshalStateInfo(rawStateInfo)
			write(" " + stateInfo.String())
		}

	}
}

func dumpRenamed(tx *bolt.Tx, write WriteStr, writeRaw WriteRaw) {
	renamed := tx.Bucket(renamedBucket)
	if renamed == nil {
		write("error: missing renamed\n")
	} else {
		write("renamed:\n")
		cursor := renamed.Cursor()
		oldName, newName := cursor.First()
		for ; oldName != nil; oldName, newName = cursor.Next() {
			write("  ")
			writeRaw(oldName)
			write(" → ")
			writeRaw(newName)
			write("\n")
		}
	}
}

func dumpSaves(tx *bolt.Tx, write WriteStr, writeRaw WriteRaw) error {
	saves := tx.Bucket(savesBucket)
	if saves == nil {
		write("error: missing saves\n")
	} else {
		write("saves:\n")
		cursor := saves.Cursor()
		rawSid, _ := cursor.First()
		for ; rawSid != nil; rawSid, _ = cursor.Next() {
			sid := UnmarshalSid(rawSid)
			write(fmt.Sprintf("  sid #%d: ", sid))
			if err := dumpSave(saves, rawSid, write, writeRaw); err != nil {
				return err
			}
		}
	}
	return nil
}

func dumpSave(saves *bolt.Bucket, rawSid []byte, write WriteStr,
	writeRaw WriteRaw) error {
	save := saves.Bucket(rawSid)
	if save == nil {
		write("error: missing save\n")
	} else {
		rawWhen := save.Get(savesWhen)
		when, err := UnmarshalTime(rawWhen)
		if err != nil {
			return err
		}
		write(when.Format(time.DateTime))
		rawComment := save.Get(savesComment)
		if len(rawComment) > 0 {
			write(" ")
			writeRaw(rawComment)
		}
		write("\n")
		filesCursor := save.Cursor()
		filename, rawEntry := filesCursor.First()
		for ; filename != nil; filename,
			rawEntry = filesCursor.Next() {
			if slices.Equal(filename, savesWhen) ||
				slices.Equal(filename, savesComment) {
				continue // these aren't filenames
			}
			dumpEntry(filename, rawEntry, write, writeRaw)
		}
	}
	return nil
}

func dumpEntry(filename []byte, rawEntry []byte, write WriteStr,
	writeRaw WriteRaw) {
	write("    ")
	writeRaw(filename)
	entry := UnmarshalEntry(rawEntry)
	write(" " + entry.String() + "\n")
}
