// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	_ "embed"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const (
	ModeOwnerRW = 0o600
)

var (
	//go:embed Version.dat
	Version string

	StateBucket   = []byte("state")
	SavesBucket   = []byte("saves")
	RenamedBucket = []byte("renamed")
)

// Open opens (and creates if necessary) a .fhd file ready for use.
// If the returned bolt.DB is not nil, call db.Close() when finished with
// it.
func Open(filename string) (*bolt.DB, error) {
	db, err := bolt.Open(filename, ModeOwnerRW, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(StateBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", StateBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(SavesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", SavesBucket,
				err)
		}
		_, err = tx.CreateBucketIfNotExists(RenamedBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s",
				RenamedBucket, err)
		}
		return nil
	})
	if err != nil {
		closeEerr := db.Close()
		if closeEerr != nil {
			return nil, errors.Join(err, closeEerr)
		}
		return nil, err
	}
	return db, nil
}
