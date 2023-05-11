// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "fmt"

type RenameVal struct {
	NewFilename string
	OldFilename string
	OldSid      SID
}

func newRenameVal(newFilename, oldFilename string, sid SID) RenameVal {
	return RenameVal{NewFilename: newFilename, OldFilename: oldFilename,
		OldSid: sid}
}

func (me RenameVal) String() string {
	return fmt.Sprintf("%s→%s#%d", me.NewFilename, me.OldFilename,
		me.OldSid)
}

func (me RenameVal) Marshal() []byte {
	rawNewFilename := []byte(me.NewFilename) // max len 64K bytes
	rawOldFilename := []byte(me.OldFilename) // max len 64K bytes
	size := uint16size + len(rawNewFilename) + uint16size +
		len(rawOldFilename) + SidSize
	raw := make([]byte, 0, size)
	raw = append(raw, MarshalUint16(uint16(len(rawNewFilename)))...)
	raw = append(raw, rawNewFilename...)
	raw = append(raw, MarshalUint16(uint16(len(rawOldFilename)))...)
	raw = append(raw, rawOldFilename...)
	return append(raw, me.OldSid.Marshal()...)
}

func UnmarshalRenameVal(raw []byte) RenameVal {
	var renameVal RenameVal
	index := uint16size
	size := int(UnmarshalUint16(raw[:index]))
	renameVal.NewFilename = string(raw[index : index+size])
	index += size
	size = int(UnmarshalUint16(raw[index : index+uint16size]))
	index += uint16size
	renameVal.OldFilename = string(raw[index : index+size])
	index += size
	renameVal.OldSid = UnmarshalSid(raw[index : index+SidSize])
	return renameVal
}