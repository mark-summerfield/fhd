// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"compress/flate"
	"compress/lzw"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/mark-summerfield/gong"
)

func getRaws(filename string, sha *shA256) ([]byte, []byte, []byte, error) {
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
	go func() { defer wg.Done(); populateSha(raw, sha) }()
	go func() { defer wg.Done(); populateFlate(raw, &rawFlate) }()
	go func() { defer wg.Done(); populateLzw(raw, &rawLzw) }()
	wg.Wait()
	return raw, rawFlate.Bytes(), rawLzw.Bytes(), nil
}

func populateSha(raw []byte, sha *shA256) {
	*sha = shA256(sha256.Sum256(raw))
}

func populateFlate(raw []byte, rawFlate *bytes.Buffer) {
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

func populateLzw(raw []byte, rawLzw *bytes.Buffer) {
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

func getExtractFilename(sid SID, filename string) string {
	dir, base := filepath.Split(filename)
	ext := filepath.Ext(base)
	base = base[:len(base)-len(ext)]
	sep := "#"
	var extracted string
	for {
		extracted = fmt.Sprintf("%s%s%s%d%s", dir, base, sep, sid, ext)
		if !gong.FileExists(extracted) {
			break
		}
		sep += "#"
	}
	return extracted
}

func copyFile(dest, source string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, gong.ModeUserRW)
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}
