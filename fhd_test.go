package fhd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/mark-summerfield/gong"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/exp/slices"
)

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
		actual = fhd.String()
		expected := fmt.Sprintf("<Fhd filename=%q format=1>", filename)
		if actual != expected {
			t.Errorf("expected String of %q, got %q", expected, actual)
		}
	}
}

func TestCompressionForSizes(t *testing.T) {
	if compression := compressionForSizes(1000, 997,
		998); compression != noCompression {
		t.Errorf("expected noCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 945,
		998); compression != flateCompression {
		t.Errorf("expected flateCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 998,
		949); compression != lzwCompression {
		t.Errorf("expected lzwCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 0,
		990); compression != noCompression {
		t.Errorf("expected noCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 990,
		0); compression != noCompression {
		t.Errorf("expected noCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 889,
		0); compression != flateCompression {
		t.Errorf("expected flateCompression, got %s", compression)
	}
	if compression := compressionForSizes(1000, 0,
		889); compression != lzwCompression {
		t.Errorf("expected lzwCompression, got %s", compression)
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
		saveResult, err := fhd.MonitorWithComment("start", file1, file2)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveResult.Sid)
		}
		expected := "start"
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
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
		saveResult, err := fhd.MonitorWithComment(expected, files...)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 1 {
			t.Errorf("expected SID of 1, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
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

			fmt.Println("actual")
			fmt.Println(normalized(actual))
			fmt.Println("expected1")
			fmt.Println(normalized(expected1))
			fmt.Println("-------------------")
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
		saveResult, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 2 {
			t.Errorf("expected SID of 2, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
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
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		saveResult, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 3 {
			t.Errorf("expected SID of 3, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
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
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		saveResult, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 4 {
			t.Errorf("expected SID of 4, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
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
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		if err = copyFile("../5/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../5")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "this is the fifth save"
		saveResult, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 5 {
			t.Errorf("expected SID of 5, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if len(saveResult.MissingFiles) != 1 {
			t.Errorf("expected one missing files, got %v",
				saveResult.MissingFiles)
		}
		missing := "ring.py"
		if !saveResult.MissingFiles.Contains(missing) {
			t.Errorf("expected missing file %q, got %v", missing,
				saveResult.MissingFiles)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 1 {
			t.Errorf("expected 1 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			if !state.Monitored {
				if state.Filename != missing {
					t.Errorf("expected unmonitored file %q, got %q",
						missing, state.Filename)
				}
				continue
			}
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		if err = copyFile("../6/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../6")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "and for the sixth save we have these"
		saveResult, err = fhd.MonitorWithComment(expected, missing)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 6 {
			t.Errorf("expected SID of 6, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 0 {
			t.Errorf("expected no saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		if err = copyFile("../7/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../7")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		expected = "Now for save number 7."
		saveResult, err = fhd.Save(expected)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 7 {
			t.Errorf("expected SID of 7, got %d", saveResult.Sid)
		}
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of %q, got %q", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 4 {
			t.Errorf("expected 4 states, got %d", len(states))
		}
		if fhd.SaveCount() != 2 {
			t.Errorf("expected 2 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
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
		if err = copyFile("../8/"+filename, filename); err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		err = os.Chdir("../8")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		fhd, _ = New(filename)
		buffer.Reset()
		saveResult, err = fhd.Rename("computer.bmp", "pc.bmp")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if saveResult.Sid != 8 {
			t.Errorf("expected SID of 8, got %d", saveResult.Sid)
		}
		expected = "renamed \"computer.bmp\" → \"pc.bmp\""
		if saveResult.Comment != expected {
			t.Errorf("expected Comment of <%q>, got <%q>", expected,
				saveResult.Comment)
		}
		if !saveResult.MissingFiles.IsEmpty() {
			t.Errorf("expected no missing files, got %v",
				saveResult.MissingFiles)
		}
		states, err = fhd.States()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(states) != 5 {
			t.Errorf("expected 5 states, got %d", len(states))
		}
		if fhd.SaveCount() != 1 {
			t.Errorf("expected 1 saved files, got %d", fhd.SaveCount())
		}
		buffer.Reset()
		for _, state := range states {
			if !state.Monitored {
				if state.Filename != "computer.bmp" {
					t.Errorf(
						"expected unmonitored \"computer.bmp\", got %q",
						state.Filename)
				}
				continue
			}
			err = fhd.ExtractForSid(state.LastSid, state.Filename, &buffer)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			raw := buffer.Bytes()
			if !compareFileWithRaw(state.Filename, raw) {
				t.Errorf("expected equal for %s", state.Filename)
			}
			buffer.Reset()
		}

		monitored, err := fhd.Monitored()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(monitored) != 4 {
			t.Errorf("expected 4 monitored files, got %d", len(monitored))
		}
		unmonitored, err := fhd.Unmonitored()
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(unmonitored) != 1 {
			t.Errorf("expected 1 unmonitored files, got %d",
				len(unmonitored))
		}

		err = os.Chdir("../1")
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		buffer.Reset()
		for _, state := range states {
			if state.Filename == "pc.bmp" {
				continue // not in first save
			}
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
	for i := 1; i < 9; i++ {
		os.Remove("tdata/" + strconv.Itoa(i) + "/" + filename)
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
  battery.png M#1:I
  computer.bmp M#1:I
  ring.py M#1:T
  wordsearch.pyw M#1:T
saves:
  sid #1: 2023-05-11 08:45:38 started
    battery.png U 2,525 bytes SHA256=7c94b6962b6f…7b6dfdb68c6 
    computer.bmp F 3,693 bytes SHA256=d274b3d4b89c…a858ff627ea 
    ring.py F 657 bytes SHA256=831de79f9c70…af3d6b9266b 
    wordsearch.pyw F 1,296 bytes SHA256=432823716ba1…123467d5a6c`
)
