package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/simplesurance/dependencies-tool/internal/deps"
)

const deployOrderShortHelp = "Generate a deployment order."

var descrDependencyFileNames = strings.TrimSpace(`
Dependency definition files are discovered by searching in all child directories
of PATH.
Files that match the following names are discovered, only the first found per
directory is parsed, the preference order is:

  1. .deps-<ENVIRONMENT>-<REGION>.toml
  2. .deps-<ENVIRONMENT>.toml
  3. .deps.toml
`)

var deployOrderLongHelp = deployOrderShortHelp + "\n\n" + strings.TrimSpace(`
Positional Arguments:
  PATH		- Root Directory for the dependency file discovery.
  ENVIRONMENT   - Value that is used as the ENVIRONMENT placeholder of the 
                  searched dependency file names.
  REGION        - Value that is used as the REGION placeholder of the searched
		  dependency file names.
  APP-NAME      - Application names for that the deployment order is generated.
                  If not specified, the dependencies of all applications are 
		  evaluated.

`+descrDependencyFileNames)

type deployOrder struct {
	*cobra.Command

	Format string

	// positional arguments
	Path   string
	Env    string
	Region string
	Apps   []string
}

func newDeployOrder() *deployOrder {
	supportedFormats := []string{"text", "dot"}

	cmd := deployOrder{
		Command: &cobra.Command{
			Use:   "deploy-order PATH ENVIRONMENT REGION [APP-NAME]...]",
			Short: "Generate a deployment order from dependencies",
			Long:  deployOrderLongHelp,
			Args:  cobra.MinimumNArgs(3),
		},
	}

	cmd.Flags().StringVar(
		&cmd.Format, "format", "text",
		fmt.Sprintf("output format, supported values: %s",
			strings.Join(supportedFormats, ", ")),
	)
	cmd.PreRunE = func(_ *cobra.Command, args []string) error {
		if !slices.Contains(supportedFormats, cmd.Format) {
			return fmt.Errorf("unsupported --format values: %q, expecting one of: %s ", cmd.Format,
				strings.Join(supportedFormats, ", "))
		}

		cmd.Path = args[0]
		cmd.Env = args[1]
		cmd.Region = args[2]
		if len(args) > 3 {
			cmd.Apps = args[3:]
		}

		return nil
	}
	cmd.RunE = cmd.run

	return &cmd
}

func (c *deployOrder) run(*cobra.Command, []string) error {
	var depsfrom deps.Composition

	composition, err := deps.CompositionFromSisuDir(c.Path, c.Env, c.Region)
	if err != nil {
		return err
	}

	if len(c.Apps) == 0 {
		depsfrom = composition
	} else {
		deps, err := composition.RecursiveDepsOf(strings.Join(c.Apps, ","))
		if err != nil {
			return err
		}
		depsfrom = *deps
	}

	switch c.Format {
	case "text":
		secondsorted, err := depsfrom.DeploymentOrder()
		if err != nil {
			return fmt.Errorf("could not generate graph: %w", err)
		}

		for _, i := range secondsorted {
			fmt.Println(i)
		}
	case "dot":
		fmt.Printf("###########\n# dot of %s\n##########\n", strings.Join(c.Apps, ", "))
		depsgraph, err := deps.OutputDotGraph(depsfrom)
		if err != nil {
			return err
		}

		fmt.Println(depsgraph)
	}

	return nil
}
