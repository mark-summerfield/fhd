// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import "crypto/sha256"

type Entry struct {
	Sha256 [sha256.Size]byte
	Flag   Flag
	Blob   []byte
}

// TODO
// func newEntry() *Entry
// func UnmarshalEntry(raw []byte) *Entry
// func (me *Entry) Marshal() []byte

// first sha256.Size bytes is Sha256, next byte is flag, remaining bytes are Blob
