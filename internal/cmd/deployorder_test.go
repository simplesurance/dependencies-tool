package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"
)

var relTestDataDirPath = filepath.Join("..", "deps", "testdata")

func TestExportImport(t *testing.T) {
	tmpdir := t.TempDir()
	destFile := filepath.Join(tmpdir, "tree.json")

	exportcmd := newExportCmd()
	exportcmd.root = relTestDataDirPath
	exportcmd.region = "eu"
	exportcmd.env = "stg"
	exportcmd.destFile = destFile

	err := exportcmd.run(exportcmd.Command, nil)
	if err != nil {
		t.Fatal(err)
	}

	deployOrderCmd := newDeployOrder()

	err = deployOrderCmd.PreRunE(nil, []string{destFile})
	if err != nil {
		t.Fatal(err)
	}

	stdoutBuf := bytes.Buffer{}
	deployOrderCmd.Command.SetOut(&stdoutBuf)
	deployOrderCmd.includeAppsWithoutDeployDir = true

	err = deployOrderCmd.run(deployOrderCmd.Command, nil)
	if err != nil {
		t.Fatal(err)
	}
	const expectedOut = `b-service
stg-eu-service
postgres
consul
a-service
`
	outStr := stdoutBuf.String()
	if outStr != expectedOut {
		t.Errorf("got deployorder:\n---\n%s---\n\nexpected:\n---\n%s\n---", outStr, expectedOut)
	}
}

func TestDeployOrderInJson(t *testing.T) {
	deployOrderCmd := newDeployOrder()
	deployOrderCmd.format = "json"
	deployOrderCmd.includeAppsWithoutDeployDir = true

	err := deployOrderCmd.PreRunE(nil, []string{relTestDataDirPath, "stg", "eu"})
	if err != nil {
		t.Fatal(err)
	}

	stdoutBuf := bytes.Buffer{}
	deployOrderCmd.Command.SetOut(&stdoutBuf)

	err = deployOrderCmd.run(deployOrderCmd.Command, nil)
	if err != nil {
		t.Fatal(err)
	}

	var res []string
	err = json.Unmarshal(stdoutBuf.Bytes(), &res)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"b-service", "stg-eu-service", "postgres", "consul", "a-service"}
	if !slices.Equal(res, expected) {
		t.Errorf("expected unmarshaled json result: %v, got: %v", expected, res)
	}
}
