// Copyright © 2023 Mark Summerfield. All rights reserved.
// License: Apache-2.0

package fhd

import (
	"bytes"
	"compress/flate"
	"compress/lzw"
	"crypto/sha256"
	"errors"
	"net/http"
	"os"
	"sync"
)

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

func getMimeType(filename string) string {
	file, err := os.Open(filename)
	if err == nil {
		defer file.Close()
		buffer := make([]byte, 512)
		_, err := file.Read(buffer)
		if err == nil {
			return http.DetectContentType(buffer)
		}
	}
	return "application/octet-stream"
}
