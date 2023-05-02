// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func newDb(filename string) (*bolt.DB, error) {
	db, err := bolt.Open(filename, modeUserRW, nil)
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
		_, err = tx.CreateBucketIfNotExists(stateBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", stateBucket,
				err)
		}
		saves, err := tx.CreateBucketIfNotExists(savesBucket)
		if err != nil {
			return fmt.Errorf("failed to create bucket %q: %s", savesBucket,
				err)
		}
		err = saves.SetSequence(0) // next i.e., first used, will be 1.
		return err
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

func utob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btou(b []byte) (uint64, error) {
	var u uint64
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.BigEndian, &u)
	if err != nil {
		return 0, err
	}
	return u, nil
}
