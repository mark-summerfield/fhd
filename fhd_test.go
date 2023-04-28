package fhd

import (
	"os"
	"path/filepath"
	"testing"
	// "github.com/mark-summerfield/gong"
	// "golang.org/x/exp/maps"
	// "golang.org/x/exp/slices"
)

// maps.Equal() & maps.EqualFunc() & slices.Equal() & slices.EqualFunc()
// https://pkg.go.dev/golang.org/x/exp/maps
// https://pkg.go.dev/golang.org/x/exp/slices
// gong.IsRealClose() & gong.IsRealZero()

func Test001(t *testing.T) {
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
	}
}
