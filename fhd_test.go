package fhd

import (
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
	expected := "Hello fhd v0.1.0\n"
	actual := Hello()
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}
