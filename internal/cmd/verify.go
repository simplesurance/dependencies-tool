package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

const verifyShortHelp = "Verify dependency files"

var verifyLongHelp = verifyShortHelp + "\n\n" + strings.TrimSpace(`
Positional Arguments:
`+descRootDirArg+`
`+descrEnvRegionArgs+`
`+descrDependencyFileNames)

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
			Short: verifyShortHelp,
			Long:  verifyLongHelp,
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

func (c *verify) run(cc *cobra.Command, _ []string) error {
	composition, err := deps.CompositionFromSisuDir(c.path, c.env, c.region)
	if err != nil {
		return err
	}

	if err := composition.VerifyDependencies(); err != nil {
		return err
	}

	cc.Println("verification successful")

	return nil
}
