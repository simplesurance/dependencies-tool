package cmd

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestExportImport(t *testing.T) {
	tmpdir := t.TempDir()
	destFile := filepath.Join(tmpdir, "tree.json")

	exportcmd := newExportCmd()
	exportcmd.root = filepath.Join("..", "deps", "testdata")
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

	err = deployOrderCmd.run(deployOrderCmd.Command, nil)
	if err != nil {
		t.Fatal(err)
	}
	const expectedOut = `stg-eu-service
postgres
consul
a-service
`
	outStr := stdoutBuf.String()
	if outStr != expectedOut {
		t.Errorf("got deployorder:\n---\n%s---\n\nexpected:\n---\n%s\n---", outStr, expectedOut)
	}
}
