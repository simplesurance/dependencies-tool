package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

type verify struct {
	*cobra.Command
	path   string
	env    string
	region string
}

func newVerify() *verify {
	cmd := verify{
		Command: &cobra.Command{
			Use:   "verify PATH ENVIRONMENT REGION",
			Short: "Verify dependency files",
			Args:  cobra.ExactArgs(3),
		},
	}

	cmd.RunE = cmd.run

	cmd.PreRun = func(_ *cobra.Command, args []string) {
		cmd.path = args[0]
		cmd.env = args[1]
		cmd.region = args[2]
	}

	return &cmd
}

func (c *verify) run(*cobra.Command, []string) error {
	composition, err := deps.CompositionFromSisuDir(c.path, c.env, c.region)
	if err != nil {
		return err
	}

	if err := composition.VerifyDependencies(); err != nil {
		return err
	}

	fmt.Println("verification successful")

	return nil
}
