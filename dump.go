// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"fmt"
	"io"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

type (
	writeStr func(string)
	writeRaw func([]byte)
)

func (me *Fhd) Dump() error {
	return me.DumpTo(os.Stderr)
}

// DumpTo writes data from the underlying database to the writer purely for
// debugging and testing.
func (me *Fhd) DumpTo(writer io.Writer) error {
	write := func(text string) { _, _ = writer.Write([]byte(text)) }
	writeRaw := func(raw []byte) { _, _ = writer.Write(raw) }
	return me.db.View(func(tx *bolt.Tx) error {
		dumpConfig(tx, write, writeRaw)
		dumpStates(tx, write, writeRaw)
		return dumpSaves(tx, write, writeRaw)
	})
}

func dumpConfig(tx *bolt.Tx, write writeStr, writeRaw writeRaw) {
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
		ignore := config.Bucket(configIgnore)
		if ignore == nil {
			write("  error (ignore is missing)\n")
		} else {
			write("  ignore=")
			cursor := ignore.Cursor()
			rawFilename, _ := cursor.First()
			for ; rawFilename != nil; rawFilename, _ = cursor.Next() {
				write(" \"")
				writeRaw(rawFilename)
				write("\"")
			}
			write("\n")
		}
	}
}

func dumpStates(tx *bolt.Tx, write writeStr, writeRaw writeRaw) {
	states := tx.Bucket(statesBucket)
	if states == nil {
		write("error: missing states\n")
	} else {
		write("states:\n")
		cursor := states.Cursor()
		rawFilename, rawStateVal := cursor.First()
		for ; rawFilename != nil; rawFilename, rawStateVal = cursor.Next() {
			write("  ")
			writeRaw(rawFilename)
			stateVal := unmarshalStateVal(rawStateVal)
			write(" " + stateVal.String())
			write("\n")
		}

	}
}

func dumpSaves(tx *bolt.Tx, write writeStr, writeRaw writeRaw) error {
	saves := tx.Bucket(savesBucket)
	if saves == nil {
		write("error: missing saves\n")
	} else {
		write("saves:\n")
		cursor := saves.Cursor()
		rawSid, _ := cursor.First()
		for ; rawSid != nil; rawSid, _ = cursor.Next() {
			sid := unmarshalSid(rawSid)
			write(fmt.Sprintf("  sid #%d: ", sid))
			if err := dumpSaveItem(tx, rawSid, write,
				writeRaw); err != nil {
				return err
			}
			if err := dumpSave(saves, rawSid, write, writeRaw); err != nil {
				return err
			}
		}
	}
	return nil
}

func dumpSaveItem(tx *bolt.Tx, rawSid []byte, write writeStr,
	writeRaw writeRaw) error {
	saveItems := tx.Bucket(saveItemsBucket)
	if saveItems == nil {
		write("error: missing save\n")
	} else {
		rawSaveVal := saveItems.Get(rawSid)
		if rawSaveVal == nil {
			write("error: missing saveval\n")
		} else {
			saveVal, err := unmarshalSaveVal(rawSaveVal)
			if err != nil {
				write(fmt.Sprintf("error: unmarshal saveval: %s", err))
			} else {
				write(saveVal.When.Format(time.DateTime))
				if len(saveVal.Comment) > 0 {
					write(" ")
					write(saveVal.Comment)
				}
				write("\n")
			}
		}
	}
	return nil
}

func dumpSave(saves *bolt.Bucket, rawSid []byte, write writeStr,
	writeRaw writeRaw) error {
	save := saves.Bucket(rawSid)
	if save == nil {
		write("error: missing save\n")
	} else {
		cursor := save.Cursor()
		rawFilename, rawEntry := cursor.First()
		for ; rawFilename != nil; rawFilename, rawEntry = cursor.Next() {
			dumpEntry(rawFilename, rawEntry, write, writeRaw)
		}
	}
	return nil
}

func dumpEntry(rawFilename []byte, rawEntry []byte, write writeStr,
	writeRaw writeRaw) {
	write("    ")
	writeRaw(rawFilename)
	entry := unmarshalEntry(rawEntry)
	write(" " + entry.String() + "\n")
}
