// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

type Entry struct {
	Sha256 []byte
	Flag   Flag
	Blob   []byte
}

// TODO newEntry, Marshal, Unmarshal
// first sha256.Size bytes is Sha256, next byte is flag, remaining bytes are Blob
