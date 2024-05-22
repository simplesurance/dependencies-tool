package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/v3/internal/deps"
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
			Use:   "export ROOT-DIR [DEST-FILE]",
			Short: exportCmdShortHelp,
			Long:  exportCmdLongHelp,
			Args:  cobra.RangeArgs(1, 2),
		},
	}

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		cmd.root = args[0]
		if len(args) >= 2 {
			cmd.destFile = args[1]
		}
		return nil
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *exportCmd) run(cc *cobra.Command, _ []string) error {
	cmp, err := deps.CompositionFromDir(c.root, c.rootCmd.cfgName, c.rootCmd.ignoredDirs)
	if err != nil {
		return err
	}

	if cmp.IsEmpty() {
		return fmt.Errorf("could not find any dependency information in %s", c.root)
	}

	if c.destFile == "" {
		if err := cmp.ToJSON(os.Stdout); err != nil {
			return err
		}
	} else {
		err = cmp.ToJSONFile(c.destFile)
		if err != nil {
			return err
		}
		cc.Printf("written dependency tree to %s\n", filepath.Clean(c.destFile))
	}

	return nil
}
