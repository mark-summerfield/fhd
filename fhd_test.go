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
			if config := tx.Bucket(configBucket); config == nil {
				t.Error("expected configBucket, got nil")
			}
			if states := tx.Bucket(statesBucket); states == nil {
				t.Error("expected statesBucket, got nil")
			}
			if renames := tx.Bucket(renamedBucket); renames == nil {
				t.Error("expected renamedBucket, got nil")
			}
			if saves := tx.Bucket(savesBucket); saves == nil {
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
	if flag := flagForSizes(1000, 997, 998); flag != Raw {
		t.Errorf("expected Raw, got %s", flag)
	}
	if flag := flagForSizes(1000, 945, 998); flag != Flate {
		t.Errorf("expected Flate, got %s", flag)
	}
	if flag := flagForSizes(1000, 998, 949); flag != Lzw {
		t.Errorf("expected Lzw, got %s", flag)
	}
	if flag := flagForSizes(1000, 0, 990); flag != Raw {
		t.Errorf("expected Raw, got %s", flag)
	}
	if flag := flagForSizes(1000, 990, 0); flag != Raw {
		t.Errorf("expected Raw, got %s", flag)
	}
	if flag := flagForSizes(1000, 889, 0); flag != Flate {
		t.Errorf("expected Flate, got %s", flag)
	}
	if flag := flagForSizes(1000, 0, 889); flag != Lzw {
		t.Errorf("expected Lzw, got %s", flag)
	}
}

func TestSidSequence(t *testing.T) {
	filename := filepath.Join(os.TempDir(), "temp2.fhd")
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		expected := "save #1"
		sidInfo, err := fhd.newSid(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if sidInfo.Sid != 1 {
			t.Errorf("unexpected sid, expected 1, got %d", sidInfo.Sid)
		}
		if sidInfo.Comment != expected {
			t.Errorf("unexpected sid, expected %s, got %s", expected,
				sidInfo.Comment)
		}
		err = fhd.db.Update(func(tx *bolt.Tx) error {
			saves := tx.Bucket(savesBucket)
			if saves == nil {
				err := fmt.Errorf("expected savesBucket, got nil")
				t.Error(err)
				return err
			}
			_, err := saves.CreateBucket(sidInfo.RawSid())
			if err != nil {
				t.Error(err)
				return err
			}
			u, _ := saves.NextSequence()
			sid := SID(u)
			if sid != 2 {
				t.Errorf("expected sid of 2: %d", sid)
			}
			_, err = saves.CreateBucket(MarshalSid(sidInfo.Sid + 1))
			if err != nil {
				t.Error(err)
				return err
			}
			u, _ = saves.NextSequence()
			sid = SID(u)
			if sid != 3 {
				t.Errorf("expected sid of 3: %d", sid)
			}
			_, err = saves.CreateBucket(MarshalSid(sidInfo.Sid + 2))
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
			rawSid, _ := cursor.Last()
			if rawSid == nil {
				t.Error("expected savesBucket sid 3, got nil")
			}
			sid := UnmarshalSid(rawSid)
			if sid != 3 {
				t.Errorf("expected savesBucket expected 3 got %v", sid)
			}
			rawSid, _ = cursor.Prev()
			if rawSid == nil {
				t.Error("expected savesBucket sid 2, got nil")
			}
			sid = UnmarshalSid(rawSid)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if sid != 2 {
				t.Errorf("expected savesBucket expected 2 got %v", sid)
			}
			rawSid, _ = cursor.Prev()
			if rawSid == nil {
				t.Error("expected savesBucket sid 1, got nil")
			}
			sid = UnmarshalSid(rawSid)
			if sid != 1 {
				t.Errorf("expected savesBucket expected 1 got %v", sid)
			}
			rawSid, _ = cursor.Prev()
			if rawSid != nil {
				t.Errorf("expected savesBucket nil got %v", rawSid)
			}
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	}
}
