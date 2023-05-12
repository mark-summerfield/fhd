package fhd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "temp1.fhd")
	os.Remove(filename)
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	// defer func() { os.Remove(filename) }() // TODO
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
