// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark-summerfield/gong"
)

type SHA256 [sha256.Size]byte

type Entry struct {
	Sha256 SHA256
	Flag   Flag
	Blob   []byte
}

func newEntry(sha256 SHA256, flag Flag, blob []byte) *Entry {
	return &Entry{Sha256: sha256, Flag: flag, Blob: blob}
}

func UnmarshalEntry(raw []byte) *Entry {
	return &Entry{Sha256: SHA256(raw[:sha256.Size]),
		Flag: Flag(raw[sha256.Size]), Blob: raw[sha256.Size+1:]}
}

func (me *Entry) Marshal() []byte {
	raw := make([]byte, 0, sha256.Size+1+len(me.Blob))
	raw = append(raw, me.Sha256[:]...)
	raw = append(raw, byte(me.Flag))
	return append(raw, me.Blob...)
}

// String is for Dump() and debugging.
func (me *Entry) String() string {
	var text strings.Builder
	text.WriteString(fmt.Sprintf("SHA256=%v %s ", me.Sha256, me.Flag))
	if me.Flag == Raw && strings.HasPrefix(
		http.DetectContentType(me.Blob), "text") {
		text.WriteByte('"')
		text.WriteString(gong.ElideMiddle(string(me.Blob), 40))
		text.WriteByte('"')
	} else {
		text.WriteString(gong.Commas(len(me.Blob)))
		text.WriteString("bytes")
	}
	return text.String()
}
