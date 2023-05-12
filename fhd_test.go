package fhd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	if flag := flagForSizes(1000, 997, 998); flag != rawFlag {
		t.Errorf("expected rawFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 945, 998); flag != flateFlag {
		t.Errorf("expected flateFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 998, 949); flag != lzwFlag {
		t.Errorf("expected lzwFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 0, 990); flag != rawFlag {
		t.Errorf("expected rawFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 990, 0); flag != rawFlag {
		t.Errorf("expected rawFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 889, 0); flag != flateFlag {
		t.Errorf("expected flateFlag, got %s", flag)
	}
	if flag := flagForSizes(1000, 0, 889); flag != lzwFlag {
		t.Errorf("expected lzwFlag, got %s", flag)
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
		saveItem, err := fhd.MonitorWithComment("start", file1, file2)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveItem.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveItem.Sid)
		}
		expected := "start"
		if saveItem.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveItem.Comment)
		}
	}
}

func Test_tdata(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		dir = gong.AbsPath(".")
	}
	defer func() { _ = os.Chdir(dir) }()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	} else {
		filename := "tdata.fhd"
		removeFhds(filename)
		err = os.Chdir("tdata/1")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ := New(filename)
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
		saveItem, err := fhd.MonitorWithComment(expected, files...)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveItem.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveItem.Sid)
		}
		if saveItem.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveItem.Comment)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if len(states) != fhd.SaveCount() {
			t.Errorf("expected 4 files, got %d", fhd.SaveCount())
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
		buffer.Reset()
		_ = fhd.DumpTo(&buffer)
		actual := buffer.String()
		if normalized(actual) != normalized(expected1) {
			t.Error("DumpTo doesn't match expected1")
		}

		err = fhd.Close()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if err = copyFile("../2/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../2")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "second save"
		saveItem, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveItem.Sid != 2 {
			t.Errorf("expected SID of 2, got %d", saveItem.Sid)
		}
		if saveItem.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveItem.Comment)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 3 {
			t.Errorf("expected 3 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(state.Sid, state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}

		err = fhd.Close()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if err = copyFile("../3/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../3")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "the third save"
		saveItem, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveItem.Sid != 3 {
			t.Errorf("expected SID of 3, got %d", saveItem.Sid)
		}
		if saveItem.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveItem.Comment)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 3 {
			t.Errorf("expected 3 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(state.Sid, state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}

		err = fhd.Close()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if err = copyFile("../4/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../4")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "and the fourth save"
		saveItem, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveItem.Sid != 4 {
			t.Errorf("expected SID of 4, got %d", saveItem.Sid)
		}
		if saveItem.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveItem.Comment)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 4 {
			t.Errorf("expected 4 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(state.Sid, state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}

		err = os.Chdir("../1")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(1, state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}
		err = os.Chdir(dir)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		removeFhds(filename)
	}
}

func removeFhds(filename string) {
	for _, i := range []string{"1", "2", "3", "4"} {
		os.Remove("tdata/" + i + "/" + filename)
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

func normalized(text string) string {
	rx := regexp.MustCompile(`20\d\d-\d\d-\d\d \d\d:\d\d:\d\d`)
	return rx.ReplaceAllLiteralString(strings.TrimSpace(text),
		"<TIMESTAMP>")
}

const (
	expected1 = `config
  format=1
  ignore= "*#[0-9].*" "*.a" "*.bak" "*.class" "*.dll" "*.exe" "*.fhd" "*.jar" "*.ld" "*.ldx" "*.li" "*.lix" "*.o" "*.obj" "*.py[co]" "*.rs.bk" "*.so" "*.sw[nop]" "*.swp" "*.tmp" "*~" "gpl-[0-9].[0-9].txt" "louti[0-9]*" "moc_*.cpp" "qrc_*.cpp" "ui_*.h"
states:
  battery.png M #1:I
  computer.bmp M #1:I
  ring.py M #1:T
  wordsearch.pyw M #1:T
renamed:
saves:
  sid #1: 2023-05-11 08:45:38 started
    battery.png R 2,525 bytes SHA256=7c94b6962b6f…7b6dfdb68c6 
    computer.bmp F 3,693 bytes SHA256=d274b3d4b89c…a858ff627ea 
    ring.py F 657 bytes SHA256=831de79f9c70…af3d6b9266b 
    wordsearch.pyw F 1,296 bytes SHA256=432823716ba1…123467d5a6c`
)
