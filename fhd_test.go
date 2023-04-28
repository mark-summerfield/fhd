package fhd

import (
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"
	// "github.com/mark-summerfield/gong"
	// "golang.org/x/exp/maps"
	// "golang.org/x/exp/slices"
)

// maps.Equal() & maps.EqualFunc() & slices.Equal() & slices.EqualFunc()
// https://pkg.go.dev/golang.org/x/exp/maps
// https://pkg.go.dev/golang.org/x/exp/slices
// gong.IsRealClose() & gong.IsRealZero()

func TestOpen(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "temp1.fhd")
	db, err := Open(filename)
	defer func() { _ = db.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		expected := "DB<\"/tmp/temp1.fhd\">"
		actual := db.String()
		if actual != expected {
			t.Errorf("expected %q, got %q", expected, actual)
		}
		actual = db.Path()
		if actual != filename {
			t.Errorf("expected %q, got %q", filename, actual)
		}
		err = db.View(func(tx *bolt.Tx) error {
			if buck := tx.Bucket(StateBucket); buck == nil {
				t.Error("expected StateBucket, got nil")
			}
			if buck := tx.Bucket(SavesBucket); buck == nil {
				t.Error("expected SavesBucket, got nil")
			}
			if buck := tx.Bucket(RenamedBucket); buck == nil {
				t.Error("expected RenamedBucket, got nil")
			}
			if buck := tx.Bucket([]byte{'x'}); buck != nil {
				t.Errorf("expected nil bucket, got %v", buck)
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}
}
