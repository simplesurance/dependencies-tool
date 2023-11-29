package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

type verify struct {
	*cobra.Command
	Path   string
	Env    string
	Region string
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
		cmd.Path = args[0]
		cmd.Env = args[1]
		cmd.Region = args[2]
	}

	return &cmd
}

func (c *verify) run(*cobra.Command, []string) error {
	composition, err := deps.CompositionFromSisuDir(c.Path, c.Env, c.Region)
	if err != nil {
		return err
	}

	if err := composition.VerifyDependencies(); err != nil {
		return err
	}

	fmt.Println("verification successful")

	return nil
}
