package testutils

import (
	"slices"
	"testing"
)

// SliceIdx returns the index of the element v in the slice sl.
// If the element is not found t.Fatal is called.
func SliceIdx[S ~[]E, E comparable](t *testing.T, sl S, v E) int {
	t.Helper()

	idx := slices.Index(sl, v)
	if idx == -1 {
		t.Fatalf("%v not found in %v", v, sl)
	}
	return idx
}

// After asserts that element a is ordered after element b in the slice sl.
func After[S ~[]E, E comparable](t *testing.T, sl S, a, b E) {
	t.Helper()

	aIdx := SliceIdx(t, sl, a)
	bIdx := SliceIdx(t, sl, b)

	if aIdx < bIdx {
		t.Fatalf("element %v (idx: %d) is ordered before %v (idx: %d)", a, aIdx, b, bIdx)
	}
}
