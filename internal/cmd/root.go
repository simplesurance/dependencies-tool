package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newRoot() *cobra.Command {
	const shortDesc = "Visualize Dependencies and generate deployment orders"

	root := &cobra.Command{
		Short:        shortDesc,
		SilenceUsage: true,
	}

	root.AddCommand(newDeployOrder().Command)
	root.AddCommand(newVerify().Command)
	root.AddCommand(newExportCmd().Command)

	return root
}

func Execute() {
	root := newRoot()
	root.SetOut(os.Stdout)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
