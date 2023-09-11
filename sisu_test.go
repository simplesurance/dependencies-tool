//go:build !integrationtest || unittest
// +build !integrationtest unittest

package main

import (
	"testing"
)

func TestAppTomlsFromAppDirFile(t *testing.T) {
	tomls, err := appTomlsFromAppDirFile("test/appdirfile")
	if err != nil {
		t.Errorf("expected no error reading appdirFile, got %v", err)
	}
	if len(tomls) != 1 {
		t.Errorf("expected one deps file from appdirfile, got %v", len(tomls))
	}
}

func TestCompositionFromAppdirFile(t *testing.T) {
	comp, err := compositionFromAppdirFile("test/appdirfile")
	if err != nil {
		t.Errorf("expected no error reading composition from appdirFile, got %v", err)
	}

	if len(comp.Services) != 1 {
		t.Errorf("expected 1 services in composition from appdirFile, got %v", len(comp.Services))
	}

	_, err = compositionFromAppdirFile("/tmp/nonexistent")
	if err == nil {
		t.Errorf("expected error reading composition from /tmp/nonexistent appdirFile, got %v", err)
	}
}

func TestCompositionFromSisuDir(t *testing.T) {
	comp, err := compositionFromSisuDir("test")
	if err != nil {
		t.Errorf("expected no error reading composition from sisudir test, got %v", err)
	}

	if len(comp.Services) == 0 {
		t.Errorf("expected multiple services in composition from appdirFile, got %v", len(comp.Services))
	}
}

func TestCompositionFromSisuDirenv2(t *testing.T) {

	tables := []struct {
		env string
		reg string
		exp string
	}{
		{"stg", "", "\"stg-service\""},
		{"", "eu", "\"eu-service\""},
		{"stg", "eu", "\"stg-eu-service\""},
		{"", "", "\"default-service\""},
		{"noenv", "", "\"default-service\""},
		{"", "noreg", "\"default-service\""},
		{"noenv", "noreg", "\"default-service\""},
	}

	for _, table := range tables {
		environment = table.env
		region = table.reg
		f := false

		comp, err := compositionFromSisuDir("test")
		if err != nil {
			t.Errorf("expected no error reading composition from sisudir test, got %v", err)
		}
		list := comp.Deps("a-service")
		//fmt.Println(table.env, table.reg, list)

		for _, b := range list {
			if b == table.exp {
				f = true
			}
		}
		if !f {
			t.Errorf("expected" + table.exp + "service as dependency for '" + table.env + "-" + table.reg + "'")
		}
	}
}

func TestLoadBaurTomlError(t *testing.T) {
	_, err := loadBaurToml("/tmp/nonexistent")
	if err == nil {
		t.Error("expected error loading /tmp/nonexistend baur.toml")
	}

	_, err = applicationTomls("/tmp/nonexistent")
	if err == nil {
		t.Error("expected error loading /tmp/nonexistend baur.toml")
	}
}
