// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/mark-summerfield/gong"
)

type shA256 [sha256.Size]byte

type saveVal struct {
	Sha         shA256
	Compression compression
	Blob        []byte
}

func newSaveVal(sha shA256, compression compression) *saveVal {
	return &saveVal{Sha: sha, Compression: compression}
}

func unmarshalSaveVal(raw []byte) *saveVal {
	return &saveVal{Sha: shA256(raw[:sha256.Size]),
		Compression: compression(raw[sha256.Size]),
		Blob:        raw[sha256.Size+1:]}
}

func (me *saveVal) marshal() []byte {
	raw := make([]byte, 0, sha256.Size+1+len(me.Blob))
	raw = append(raw, me.Sha[:]...)
	raw = append(raw, byte(me.Compression))
	return append(raw, me.Blob...)
}

// String is for Dump() and debugging.
func (me *saveVal) String() string {
	var text strings.Builder
	text.WriteString(fmt.Sprintf("%s ", me.Compression))
	if me.Compression == noCompression && strings.HasPrefix(
		http.DetectContentType(me.Blob), "text") {
		text.WriteByte('"')
		text.WriteString(gong.ElideMiddle(string(me.Blob), 32))
		text.WriteByte('"')
	} else {
		text.WriteString(gong.Commas(len(me.Blob)))
		text.WriteString(" bytes")
	}
	text.WriteString(fmt.Sprintf(" SHA256=%s ", gong.ElideMiddle(
		hex.EncodeToString(me.Sha[:]), 24)))
	return text.String()
}
