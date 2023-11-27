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
	format = "text"
	sisuDir = ""
	expectedErrMsgPrefix := "You need to define sisu directory"

	err := validateParams()
	if err == nil {
		t.Fatal("expected a validation error, got nil")
	}
	if !strings.HasPrefix(err.Error(), expectedErrMsgPrefix) {
		t.Errorf("expected errorstring in paramsvalidation to start with %v, got %v", expectedErrMsgPrefix, err.Error())
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
