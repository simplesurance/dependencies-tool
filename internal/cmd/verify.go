package cmd

import (
	"strings"

	"github.com/simplesurance/dependencies-tool/v3/internal/deps"

	"github.com/spf13/cobra"
)

const verifyShortHelp = "Verify dependency files"

var verifyLongHelp = verifyShortHelp + "\n\n" + strings.TrimSpace(`
Positional Arguments:
`+descRootDirArg,
)

type verify struct {
	*cobra.Command
	root *rootCmd
	path string
}

func newVerify(root *rootCmd) *verify {
	cmd := verify{
		root: root,
		Command: &cobra.Command{
			Use:   "verify PATH",
			Short: verifyShortHelp,
			Long:  verifyLongHelp,
			Args:  cobra.ExactArgs(1),
		},
	}

	cmd.RunE = cmd.run

	cmd.PreRun = func(_ *cobra.Command, args []string) {
		cmd.path = args[0]
	}

	return &cmd
}

func (c *verify) run(cc *cobra.Command, _ []string) error {
	_, err := deps.CompositionFromDir(c.path, c.root.cfgName, c.root.ignoredDirs)
	if err != nil {
		return err
	}

	cc.Println("verification successful, no issues found")

	return nil
}
