package cmd

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/v3/internal/cmd/fs"
)

const orderShortHelp = "Generate a deployment order."

var descrDependencyFileNames = strings.TrimSpace(`
Dependency yaml configuration files are discovered by searching in all child
directories of ROOT-DIR. Files that match the path suffix specified via
--cfg-name are parsed. Symlinks in ROOT-DIR are not followed.
`)

const descRootDirArg = `  ROOT-DIR	- Parent directory in which dependency configuration files are discovered.`

// TODO: UPDATE DESCRIPTION!
var orderLongHelp = orderShortHelp + "\n\n" + strings.TrimSpace(`
Positional Arguments:
`+descRootDirArg+`
  DEP-TREE-FILE	- Path to an exported dependency tree.

The command can use as input either a marshalled dependency-tree file (DEP-TREE-FILE)
or read and parse YAML configuration files that are found in the child directories of
ROOT-DIR to generate a dependency-tree.

`+descrDependencyFileNames)

type orderCmd struct {
	root *rootCmd
	*cobra.Command

	format string
	apps   []string

	src     string
	distr   string
	srcType fs.PathType
}

func newOrderCmd(root *rootCmd) *orderCmd {
	cmd := orderCmd{
		root: root,
		Command: &cobra.Command{
			Use:   "order ROOT-DIR|DEP-TREE-FILE DISTRIBUTION",
			Short: "Generate a dependency order",
			Long:  orderLongHelp,
			Args:  cobra.ExactArgs(2),
		},
	}

	supportedFormats := []string{"text", "dot", "json"}
	cmd.Flags().StringVar(
		&cmd.format, "format", "text",
		fmt.Sprintf("output format, supported values: %s",
			strings.Join(supportedFormats, ", ")),
	)
	cmd.Flags().StringSliceVar(
		&cmd.apps, "apps", nil,
		"comma-separated list of apps to generate the deploy order for,\n"+
			"if unset the dependency order is generated for all found apps.",
	)

	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		if !slices.Contains(supportedFormats, cmd.format) {
			return fmt.Errorf("unsupported --format values: %q, expecting one of: %s ", cmd.format,
				strings.Join(supportedFormats, ", "))
		}

		pType, err := fs.FileOrDir(args[0])
		if err != nil {
			return err
		}

		cmd.src = args[0]
		cmd.srcType = pType
		cmd.distr = args[1]

		return validateAppsParam(cmd.apps)
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *orderCmd) run(cc *cobra.Command, _ []string) error {
	composition, err := c.root.loadComposition(c.srcType, c.src)
	if err != nil {
		return err
	}

	switch c.format {
	case "text":
		order, err := composition.DependencyOrder(c.distr, c.apps...)
		if err != nil {
			return err
		}
		cc.Println(strings.Join(order, "\n"))
	case "dot":
		depsgraph, err := composition.DependencyOrderDot(c.distr, c.apps...)
		if err != nil {
			return err
		}

		cc.Print(depsgraph) // depsgraph already contains a newline at the end
	case "json":
		order, err := composition.DependencyOrder(c.distr, c.apps...)
		if err != nil {
			return err
		}
		enc := json.NewEncoder(cc.OutOrStdout())
		enc.SetIndent("", "    ")
		return enc.Encode(order)
	}

	return nil
}

func validateAppsParam(apps []string) error {
	for i, app := range apps {
		if strings.TrimSpace(app) == "" {
			return fmt.Errorf("app parameter %d contains only whitespaces or is empty: %q", i+1, app)
		}
	}
	return nil
}
