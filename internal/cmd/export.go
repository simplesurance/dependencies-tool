package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/v2/internal/deps"
)

const exportCmdShortHelp = "Read dependency definitions and export them to a file."

var exportCmdLongHelp = exportCmdShortHelp + "\n\n" +
	`Positional Arguments:
` + descRootDirArg + `
` + descrDependencyFileNames

type exportCmd struct {
	*cobra.Command
	rootCmd *rootCmd

	root     string
	destFile string
}

func newExportCmd(root *rootCmd) *exportCmd {
	cmd := exportCmd{
		rootCmd: root,
		Command: &cobra.Command{
			Use:   "export ROOT-DIR DEST-FILE",
			Short: exportCmdShortHelp,
			Long:  exportCmdLongHelp,
			Args:  cobra.ExactArgs(2),
		},
	}

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		cmd.root = args[0]
		cmd.destFile = args[1]
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

	cmp, err := deps.CompositionFromDir(c.root, c.rootCmd.cfgName, c.rootCmd.ignoredDirs)
	if err != nil {
		return err
	}

	if cmp.IsEmpty() {
		return fmt.Errorf("could not find any dependency information in %s", c.root)
	}

	err = cmp.ToJSONFile(c.destFile)
	if err != nil {
		return err
	}

	cc.Printf("written dependency tree to %s\n", absPath)

	return nil
}
