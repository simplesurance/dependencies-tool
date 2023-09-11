//go:build !integrationtest || unittest
// +build !integrationtest unittest

package main

import (
	"strings"
	"testing"
)

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestSanitize(t *testing.T) {
	s := "mystring"
	got := sanitize(s)
	exp := "\"mystring\""

	if got != exp {
		t.Errorf("expected to be result %v , got %v", exp, got)
	}
}

func TestIsIn(t *testing.T) {
	testslice := []string{"one", "two"}

	if !stringsliceContain(testslice, "one") {
		t.Error("expected to be 'one' in testslice")
	}

	if stringsliceContain(testslice, "three") {
		t.Error("three is not in testslice")
	}
}

func TestValidateParams(t *testing.T) {
	var tests = []struct {
		param    []string // input params sisu, compose-file, appdir
		expected string   //
	}{
		{[]string{"given-sisu-dir", "given-compose-file", ""}, "You can only define one of docker-compose"},
		{[]string{"given-sisu-dir", "", "given-appdir-file"}, "You can only define one of sisu directory"},
		{[]string{"", "given-compose-file", "given-appdir-file"}, "You can only define one of appdir file"},
		{[]string{"", "", ""}, "You need to define one of"},
	}
	format = "text"

	for _, tt := range tests {
		sisuDir = tt.param[0]
		composeFile = tt.param[1]
		appdirFile = tt.param[2]

		err := validateParams()
		if !strings.HasPrefix(err.Error(), tt.expected) {
			t.Errorf("expected errorstring in paramsvalidation to start with %v, got %v", tt.expected, err.Error())
		}
	}
	sisuDir = "given-sisu"
	if err := validateParams(); err != nil {
		t.Errorf("expected no error with given sisu dir, got %v", err.Error())
	}

	format = "forbidden"
	if err := validateParams(); err == nil {
		t.Error("expected error for forbidden format")
	}
}
