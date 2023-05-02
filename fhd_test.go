package fhd

import (
	"fmt"
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

func TestTest(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "temp2.fhd")
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		expected := "save #1"
		sidInfo, err := fhd.nextSid(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if sidInfo.Sid() != 1 {
			t.Errorf("unexpected sid, expected 1, got %d", sidInfo.Sid())
		}
		if sidInfo.Comment() != expected {
			t.Errorf("unexpected sid, expected %s, got %s", expected,
				sidInfo.Comment())
		}
		err = fhd.db.Update(func(tx *bolt.Tx) error {
			saves := tx.Bucket(savesBucket)
			if saves == nil {
				err := fmt.Errorf("expected savesBucket, got nil")
				t.Error(err)
				return err
			}
			_, err := saves.CreateBucket(utob(sidInfo.Sid()))
			if err != nil {
				t.Error(err)
				return err
			}
			sid, _ := saves.NextSequence()
			if sid != 2 {
				t.Errorf("expected sid of 2: %d", sid)
			}
			_, err = saves.CreateBucket(utob(sidInfo.Sid() + 1))
			if err != nil {
				t.Error(err)
				return err
			}
			sid, _ = saves.NextSequence()
			if sid != 3 {
				t.Errorf("expected sid of 3: %d", sid)
			}
			_, err = saves.CreateBucket(utob(sidInfo.Sid() + 2))
			if err != nil {
				t.Error(err)
				return err
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = fhd.db.View(func(tx *bolt.Tx) error {
			saves := tx.Bucket(savesBucket)
			if saves == nil {
				t.Error("expected savesBucket, got nil")
			}
			cursor := saves.Cursor()
			sid, _ := cursor.Last()
			if sid == nil {
				t.Error("expected savesBucket sid 3, got nil")
			}
			u, err := btou(sid)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if u != 3 {
				t.Errorf("expected savesBucket expected 3 got %v", sid)
			}
			sid, _ = cursor.Prev()
			if sid == nil {
				t.Error("expected savesBucket sid 2, got nil")
			}
			u, err = btou(sid)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if u != 2 {
				t.Errorf("expected savesBucket expected 2 got %v", sid)
			}
			sid, _ = cursor.Prev()
			if sid == nil {
				t.Error("expected savesBucket sid 1, got nil")
			}
			u, err = btou(sid)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if u != 1 {
				t.Errorf("expected savesBucket expected 1 got %v", sid)
			}
			sid, _ = cursor.Prev()
			if sid != nil {
				t.Errorf("expected savesBucket nil got %v", sid)
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}
}
