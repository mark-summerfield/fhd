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
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		actual := fhd.Filename()
		if actual != filename {
			t.Errorf("expected %q, got %q", filename, actual)
		}
		err = fhd.db.View(func(tx *bolt.Tx) error {
			if buck := tx.Bucket(configBucket); buck == nil {
				t.Error("expected configBucket, got nil")
			}
			if buck := tx.Bucket(stateBucket); buck == nil {
				t.Error("expected stateBucket, got nil")
			}
			if buck := tx.Bucket(savesBucket); buck == nil {
				t.Error("expected savesBucket, got nil")
			}
			if buck := tx.Bucket([]byte{'x'}); buck != nil {
				t.Errorf("expected nil bucket, got %v", buck)
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fileformat, err := fhd.FileFormat()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		} else if fileformat != int(fileFormat) {
			t.Errorf("unexpected file format: got %d, expected %d",
				fileformat, fileFormat)
		}
	}
}

func TestFlagForSizes(t *testing.T) {
	if flag := flagForSizes(1000, 600, 0); flag != Gz {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 970, 0); flag != Raw {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 600, 700); flag != Gz {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 970, 900); flag != Raw {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 600, 500); flag != Patch {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 970, 800); flag != Patch {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 600, 400); flag != Patch {
		t.Errorf("expected Gz, got %v", flag)
	}
	if flag := flagForSizes(1000, 970, 800); flag != Patch {
		t.Errorf("expected Gz, got %v", flag)
	}
}
