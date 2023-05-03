// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/mark-summerfield/gong"
	"golang.org/x/exp/slices"

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

// Dump writes data from the underlying database to the writer. This is for
// debugging and testing.
func (me *Fhd) Dump(writer io.Writer) error {
	writeRaw := func(raw []byte) { _, _ = writer.Write(raw) }
	write := func(text string) { _, _ = writer.Write([]byte(text)) }
	return me.db.View(func(tx *bolt.Tx) error {
		config := tx.Bucket(configBucket)
		if config == nil {
			write("error: missing config\n")
		} else {
			format := config.Get(configFormat)
			write("config/format=")
			if len(format) == 0 {
				write("nil [error]")
			} else {
				writeRaw(format)
			}
			write("\n")
		}
		states := tx.Bucket(statesBucket)
		if config == nil {
			write("error: missing states\n")
		} else {
			write("states:\n")
			cursor := states.Cursor()
			filename, state := cursor.First()
			for ; filename != nil; filename, state = cursor.Next() {
				write("  ")
				writeRaw(filename)
				kind := StateKind(state)
				write(" " + kind.String())
			}

		}
		// TODO renamed
		saves := tx.Bucket(savesBucket)
		if saves == nil {
			write("error: missing saves\n")
		} else {
			write("saves:\n")
			cursor := saves.Cursor()
			sid, _ := cursor.First()
			for ; sid != nil; sid, _ = cursor.Next() {
				u, err := btou(sid)
				if err != nil {
					return err
				}
				write(fmt.Sprintf("  sid #%d: ", u))
				save := saves.Bucket(sid)
				if save == nil {
					write("error: missing save\n")
				} else {
					rawWhen := save.Get(savesWhen)
					when, err := timeForRaw(rawWhen)
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
					filename, data := filesCursor.First()
					for ; filename != nil; filename,
						data = filesCursor.Next() {
						if slices.Equal(filename, savesWhen) ||
							slices.Equal(filename, savesComment) {
							continue // these aren't filenames
						}
						write("    ")
						writeRaw(filename)
						write(fmt.Sprintf(" datalen=%d", len(data)))
						// TODO unmarshall data: sha256 flag blob
						// write flag (as a string), sha256, elided blob if
						// mimetype HasPrefix "text/" (see
						// net/http.DetectContentType())
						write("\n")
					}
				}
			}
		}
		return nil
	})
}
