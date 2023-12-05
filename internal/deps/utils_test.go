package deps

import (
	"slices"
	"strings"
	"testing"
)

func cmpSlice(t *testing.T, expected, actual []string) {
	t.Helper()
	if !slices.Equal(expected, actual) {
		t.Errorf("got: %s, expected: %s",
			strings.Join(expected, ", "),
			strings.Join(actual, ", "),
		)
	}
}

func fatalOnErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
