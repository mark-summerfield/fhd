// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"net/http"
	"strings"
)

const (
	binKind fileKind = 'B'
	imgKind fileKind = 'I'
	txtKind fileKind = 'T'
)

type fileKind byte

func (me fileKind) String() string {
	return string(me)
}

func fileKindForRaw(raw []byte) fileKind {
	mimeType := http.DetectContentType(raw)
	if strings.HasPrefix(mimeType, "image/") {
		return imgKind
	}
	if strings.HasPrefix(mimeType, "text/") {
		return txtKind
	}
	return binKind
}
