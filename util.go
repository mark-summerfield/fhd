// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"encoding/binary"
	"time"
)

func utob(u uint64) []byte {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint64(raw, u)
	return raw
}

func btou(raw []byte) (uint64, error) {
	var u uint64
	buf := bytes.NewReader(raw)
	err := binary.Read(buf, binary.BigEndian, &u)
	if err != nil {
		return 0, err
	}
	return u, nil
}

// rawForTime is: time.Time.MarshalBinary()
func timeForRaw(raw []byte) (time.Time, error) {
	var t time.Time
	err := t.UnmarshalBinary(raw)
	return t, err
}
