// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"compress/flate"
	"compress/lzw"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"os"
	"sync"
	"time"
)

func MarshalSid(sid SID) []byte {
	raw := make([]byte, 4)
	binary.BigEndian.PutUint32(raw, uint32(sid))
	return raw
}

func UnmarshalSid(raw []byte) SID {
	var sid SID
	buf := bytes.NewReader(raw)
	err := binary.Read(buf, binary.BigEndian, &sid)
	if err != nil {
		return 0
	}
	return sid
}

// Marshal for Time is: time.Time.MarshalBinary()
func UnmarshalTime(raw []byte) (time.Time, error) {
	var t time.Time
	err := t.UnmarshalBinary(raw)
	return t, err
}

func getRaws(filename string, sha *SHA256) ([]byte, []byte, []byte, error) {
	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, err
	}
	var (
		rawFlate bytes.Buffer
		rawLzw   bytes.Buffer
		wg       sync.WaitGroup
	)
	wg.Add(3)
	go getSha(raw, sha)
	go getFlate(raw, &rawFlate)
	go getLzw(raw, &rawLzw)
	wg.Wait()
	return raw, rawFlate.Bytes(), rawLzw.Bytes(), nil
}

func getSha(raw []byte, sha *SHA256) { *sha = SHA256(sha256.Sum256(raw)) }

func getFlate(raw []byte, rawFlate *bytes.Buffer) {
	writer, err := flate.NewWriter(rawFlate, 9)
	if err == nil {
		_, ierr := writer.Write(raw)
		if ierr != nil {
			err = errors.Join(err, ierr)
		}
		ierr = writer.Close()
		if ierr != nil {
			err = errors.Join(err, ierr)
		}
		if err != nil {
			rawFlate.Reset()
		}
	}
}

func getLzw(raw []byte, rawLzw *bytes.Buffer) {
	writer := lzw.NewWriter(rawLzw, lzw.MSB, 8)
	_, err := writer.Write(raw)
	if err == nil {
		ierr := writer.Close()
		if ierr != nil {
			err = errors.Join(err, ierr)
		}
		if err != nil {
			rawLzw.Reset()
		}
	}
}
