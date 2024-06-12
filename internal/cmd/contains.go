package cmd

import (
	"fmt"
	"strings"

	"github.com/simplesurance/dependencies-tool/v3/internal/cmd/fs"

	"github.com/spf13/cobra"
)

var containsLongHelp = fmt.Sprintf(`
Check if an app is part of a distribution.

Exit Codes:
 %d - Success, app is part of the distribution
 %d - Error
 %d - App is not part of the distribution
`, ExitCodeSuccess, ExitCodeError, ExitCodeNotFound)

type containsCmd struct {
	root *rootCmd
	*cobra.Command

	src          string
	distribution string
	app          string
	srcType      fs.PathType
}

func newContainsCmd(root *rootCmd) *containsCmd {
	cmd := containsCmd{
		Command: &cobra.Command{
			Use:   "contains ROOT-DIR|DEP-TREE-FILE DISTRIBUTION APP",
			Short: "Check if an app is part of a distribution",
			Args:  cobra.ExactArgs(3),
			Long:  strings.TrimSpace(containsLongHelp),
		},
		root: root,
	}

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		cmd.src = args[0]
		cmd.distribution = args[1]
		cmd.app = args[2]

		pType, err := fs.FileOrDir(args[0])
		if err != nil {
			return err
		}

		cmd.srcType = pType

		return nil
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *containsCmd) run(*cobra.Command, []string) error {
	composition, err := c.root.loadComposition(c.srcType, c.src)
	if err != nil {
		return err
	}
	exists, err := composition.Contains(c.distribution, c.app)
	if err != nil {
		return fmt.Errorf("checking if %q is part of %q failed: %w", c.app, c.distribution, err)
	}

	if exists {
		fmt.Printf("%q is part of the distribution %q\n", c.app, c.distribution)
		return nil
	}

	fmt.Printf("%q is not part of the distribution %q\n", c.app, c.distribution)

	// do not print the error, result message has already been printed to
	// stdout
	c.SilenceErrors = true
	return NewErrWithExitCode(nil, ExitCodeNotFound)
}
