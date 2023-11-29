package cmd

import (
	"github.com/spf13/cobra"
)

func newRoot() *cobra.Command {
	const shortDesc = "Visualize Dependencies and generate deployment orders"

	root := &cobra.Command{
		Short:        shortDesc,
		Long:         descrDependencyFileNames,
		SilenceUsage: true,
	}

	root.AddCommand(newDeployOrder().Command)
	root.AddCommand(newVerify().Command)
	root.AddCommand(newExportCmd().Command)

	return root
}

func Execute() {
	root := newRoot()
	_ = root.Execute()
}
