// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"encoding/binary"
)

const (
	InvalidSID = 0
	sidSize    = 4 // *must* match type SID's size
)

type SID uint32 // allows for 4 billion saves

func (me SID) IsValid() bool { return me != InvalidSID }

func (me SID) marshal() []byte {
	raw := make([]byte, sidSize)
	binary.BigEndian.PutUint32(raw, uint32(me))
	return raw
}

func unmarshalSid(raw []byte) SID {
	var sid SID
	buf := bytes.NewReader(raw)
	if err := binary.Read(buf, binary.BigEndian, &sid); err != nil {
		return 0
	}
	return sid
}
