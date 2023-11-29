package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/cmd/fs"
	"github.com/simplesurance/dependencies-tool/internal/deps"
)

const deployOrderShortHelp = "Generate a deployment order."

var descrDependencyFileNames = strings.TrimSpace(`
The command can use as input either a marshalled dependency-tree file 
or read and parse .deps*.toml file that are found in a parent directory to
generate a dependency-tree.

If the path to a directory is specified, dependency definitions files are
discovered by searching in all child directories of SRC-PATH.
Files that match the following names are discovered, only the first found per
directory is parsed, the preference order is:

  1. .deps-<ENV>-<REGION>.toml
  2. .deps-<REGION>.toml
  3. .deps-<ENVIRONMENT>.toml
  4. .deps.toml
`)

var deployOrderLongHelp = deployOrderShortHelp + "\n\n" + strings.TrimSpace(`
Positional Arguments:
  ROOT-DIR	- Path to root directory for the dependency file discovery.
  DEP-TREE-FILE	- Path to an exported JSON dependency tree.
  ENVIRONMENT   - Value that is used as the ENVIRONMENT placeholder of the 
                  searched dependency file names.
  REGION        - Value that is used as the REGION placeholder of the searched
		  dependency file names.

`+descrDependencyFileNames)

type deployOrder struct {
	*cobra.Command

	format string
	apps   []string

	src    string
	Env    string
	region string

	srcType fs.PathType
}

func newDeployOrder() *deployOrder {
	supportedFormats := []string{"text", "dot"}

	cmd := deployOrder{
		Command: &cobra.Command{
			Use:   "deploy-order (ROOT-DIR ENVIRONMENT REGION)|DEP-TREE-FILE)",
			Short: "Generate a deployment order from dependencies",
			Long:  deployOrderLongHelp,
			Args:  cobra.MinimumNArgs(1),
		},
	}

	cmd.Flags().StringVar(
		&cmd.format, "format", "text",
		fmt.Sprintf("output format, supported values: %s",
			strings.Join(supportedFormats, ", ")),
	)
	cmd.Flags().StringSliceVar(
		&cmd.apps, "apps", nil,
		"comma-separated list of apps to generate the deploy order for, "+
			"if unset, the deploy-order is generated for all found apps.",
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

		switch pType {
		case fs.PathTypeDir:
			if len(args) != 3 {
				return fmt.Errorf("expecting 3 arguments, got: %d", len(args))
			}

			cmd.Env = args[1]
			cmd.region = args[2]

		case fs.PathTypeFile:
			if len(args) != 1 {
				return fmt.Errorf("expecting 1 arguments, got: %d", len(args))
			}

		default:
			panic(fmt.Sprintf("fileOrDir returned unexpected result (%d, %s)", pType, err))
		}

		cmd.src = args[0]
		cmd.srcType = pType

		return validateAppsParam(cmd.apps)
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *deployOrder) run(cc *cobra.Command, _ []string) error {
	var depsfrom deps.Composition

	composition, err := c.loadComposition()
	if err != nil {
		return err
	}

	if len(c.apps) == 0 {
		depsfrom = *composition
	} else {
		deps, err := composition.RecursiveDepsOf(strings.Join(c.apps, ","))
		if err != nil {
			return err
		}
		depsfrom = *deps
	}

	switch c.format {
	case "text":
		secondsorted, err := depsfrom.DeploymentOrder()
		if err != nil {
			return fmt.Errorf("could not generate graph: %w", err)
		}

		for _, i := range secondsorted {
			cc.Println(i)
		}
	case "dot":
		cc.Printf("###########\n# dot of %s\n##########\n", strings.Join(c.apps, ", "))
		depsgraph, err := deps.OutputDotGraph(depsfrom)
		if err != nil {
			return err
		}

		cc.Println(depsgraph)
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

func (c *deployOrder) loadComposition() (*deps.Composition, error) {
	switch c.srcType {
	case fs.PathTypeDir:
		return deps.CompositionFromSisuDir(c.src, c.Env, c.region)

	case fs.PathTypeFile:
		return deps.CompositionFromJSON(c.src)

	default:
		panic(fmt.Sprintf("SrcType has unexpected value: %d", c.srcType))
	}
}
