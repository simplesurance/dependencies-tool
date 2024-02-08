package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simplesurance/dependencies-tool/v3/internal/testutils"
)

var relTestDataDirPath = filepath.Join("..", "deps", "testdata")

func TestExportImport(t *testing.T) {
	tmpdir := t.TempDir()
	destFile := filepath.Join(tmpdir, "tree.json")

	// 1. export the dependency definition to destFile
	cmd := newRoot()
	cmd.SetArgs([]string{"--cfg-name", "deps.yaml", "export", relTestDataDirPath, destFile})
	err := cmd.Execute()
	require.NoError(t, err, "export command failed")

	// 2. generate a deploy order from the exported definition
	cmd = newRoot()
	cmd.SetArgs([]string{"order", "--format", "json", destFile, "prd"})
	stdoutBuf := bytes.Buffer{}
	cmd.SetOut(&stdoutBuf)
	err = cmd.Execute()
	require.NoError(t, err, "order cmd failed")

	var res []string
	err = json.Unmarshal(stdoutBuf.Bytes(), &res)
	require.NoError(t, err)

	assert.Contains(t, res, "c-service")
	testutils.After(t, res, "a-service", "b-service")

	assert.Len(t, res, 3)
}

func TestDeployOrderTextFormat(t *testing.T) {
	stdoutBuf := bytes.Buffer{}
	cmd := newRoot()
	cmd.SetArgs([]string{"order", "--cfg-name", "deps.yaml", "--format", "text", relTestDataDirPath, "prd"})
	cmd.SetOut(&stdoutBuf)
	err := cmd.Execute()
	require.NoError(t, err, "order cmd failed")

	const expected = `b-service
c-service
a-service
`
	require.Equal(t, expected, stdoutBuf.String())
}
