package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

const exportCmdShortHelp = "Generate a dependency tree and export it to a file."

var exportCmdLongHelp = exportCmdShortHelp + "\n\n" +
	`Positional Arguments:
` + descRootDirArg + `
` + descrEnvRegionArgs + `
` + descrDependencyFileNames

type exportCmd struct {
	*cobra.Command

	root     string
	env      string
	region   string
	destFile string
}

func newExportCmd() *exportCmd {
	cmd := exportCmd{
		Command: &cobra.Command{
			Use:   "export ROOT-DIR ENVIRONMENT REGION DEST-FILE",
			Short: exportCmdShortHelp,
			Long:  exportCmdLongHelp,
			Args:  cobra.ExactArgs(4),
		},
	}

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		cmd.root = args[0]
		cmd.env = args[1]
		cmd.region = args[2]
		cmd.destFile = args[3]

		return nil
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *exportCmd) run(cc *cobra.Command, _ []string) error {
	absPath, err := filepath.Abs(c.destFile)
	if err != nil {
		return err
	}

	cmp, err := deps.CompositionFromSisuDir(c.root, c.env, c.region)
	if err != nil {
		return err
	}

	err = cmp.ToJSONFile(c.destFile)
	if err != nil {
		return err
	}

	cc.Printf("written dependency tree to %s\n", absPath)

	return nil
}
