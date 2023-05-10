package fhd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark-summerfield/gong"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/exp/slices"
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
		err = fhd.db.Update(func(tx *bolt.Tx) error {
			sidInfo, err := fhd.newSid(tx, expected)
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
			saves := tx.Bucket(savesBucket)
			if saves == nil {
				err := fmt.Errorf("expected savesBucket, got nil")
				t.Error(err)
				return err
			}
			_, err = saves.CreateBucket(sidInfo.Sid.Marshal())
			if err != nil {
				t.Error(err)
				return err
			}
			u, _ := saves.NextSequence()
			sid := SID(u)
			if sid != 2 {
				t.Errorf("expected sid of 2: %d", sid)
			}
			_, err = saves.CreateBucket((sidInfo.Sid + 1).Marshal())
			if err != nil {
				t.Error(err)
				return err
			}
			u, _ = saves.NextSequence()
			sid = SID(u)
			if sid != 3 {
				t.Errorf("expected sid of 3: %d", sid)
			}
			_, err = saves.CreateBucket((sidInfo.Sid + 2).Marshal())
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

func Test2(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		dir = gong.AbsPath(".")
	}
	_ = os.Chdir(os.TempDir())
	defer func() { _ = os.Chdir(dir) }()
	filename := "temp2.fhd"
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		states, err := fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) > 0 {
			t.Errorf("expected 0 states, got %d", len(states))
		}
		file1 := "file1.txt"
		closer, err := makeTempFile(file1, "This is file1\nLine 2\n")
		defer closer()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		file2 := "file2.txt"
		closer, err = makeTempFile(file2, "This is file2\nMore\nAnd more\n")
		defer closer()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		saveInfo, err := fhd.MonitorWithComment("start", file1, file2)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveInfo.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveInfo.Sid)
		}
		expected := "start"
		if saveInfo.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveInfo.Comment)
		}
	}
}

func Test_tdata(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		dir = gong.AbsPath(".")
	}
	defer func() { _ = os.Chdir(dir) }()
	_ = os.Chdir("tdata/1")
	filename := "tdata.fhd"
	fhd, err := New(filename)
	defer func() { _ = fhd.Close() }()
	defer func() { os.Remove(filename) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		states, err := fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) > 0 {
			t.Errorf("expected 0 states, got %d", len(states))
		}
		files := []string{"battery.png", "computer.bmp", "ring.py",
			"wordsearch.pyw"}
		expected := "started"
		saveInfo, err := fhd.MonitorWithComment(expected, files...)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveInfo.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveInfo.Sid)
		}
		if saveInfo.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveInfo.Comment)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		var buffer bytes.Buffer
		for _, state := range states {
			err = fhd.Extract(state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}
		// TODO
		// cp tdata.fhd ../2
		// cd ../2
		// do work e.g., save & extract to check tdata.fhd
		// cp tdata.fhd ../3
		// cd ../3
		// do work e.g., save & extract to check tdata.fhd
		// cp tdata.fhd ../4
		// cd ../4
		// do work e.g., save & extract to check tdata.fhd
	}
}

func compareFileWithRaw(filename string, raw []byte) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		return false
	}
	return slices.Equal(data, raw)
}

func makeTempFile(filename, content string) (func(), error) {
	err := os.WriteFile(filename, []byte(content), gong.ModeUserRW)
	if err != nil {
		return func() {}, err
	}
	return func() { os.Remove(filename) }, nil
}
