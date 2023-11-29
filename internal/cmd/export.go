package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

type exportCmd struct {
	*cobra.Command

	Root     string
	Env      string
	Region   string
	DestFile string
}

func newExportCmd() *exportCmd {
	cmd := exportCmd{
		Command: &cobra.Command{
			Use:   "export ROOT-PATH ENVIRONMENT REGION DEST-FILE",
			Short: "Write the parsed dependency tree to a file.",
			Args:  cobra.ExactArgs(4),
		},
	}

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		cmd.Root = args[0]
		cmd.Env = args[1]
		cmd.Region = args[2]
		cmd.DestFile = args[3]

		return nil
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *exportCmd) run(*cobra.Command, []string) error {
	absPath, err := filepath.Abs(c.DestFile)
	if err != nil {
		return err
	}

	cmp, err := deps.CompositionFromSisuDir(c.Root, c.Env, c.Region)
	if err != nil {
		return err
	}

	err = cmp.ToJSONFile(c.DestFile)
	if err != nil {
		return err
	}

	fmt.Printf("written dependency tree to %s\n", absPath)

	return nil
}
