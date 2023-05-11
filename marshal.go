// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"encoding/binary"
	"time"
)

// Marshal for Time is: time.Time.MarshalBinary()
func unmarshalTime(raw []byte) (time.Time, error) {
	var t time.Time
	err := t.UnmarshalBinary(raw)
	return t, err
}

func marshalUint16(u uint16) []byte {
	raw := make([]byte, uint16size)
	binary.BigEndian.PutUint16(raw, u)
	return raw
}

func unmarshalUint16(raw []byte) uint16 {
	var u uint16
	buf := bytes.NewReader(raw)
	if err := binary.Read(buf, binary.BigEndian, &u); err != nil {
		return 0
	}
	return u
}
