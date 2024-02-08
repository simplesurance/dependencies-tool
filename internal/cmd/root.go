package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// version is set via goreleaser
var version = "UNDEFINED"

var defaultExcludeDirs = []string{
	".git",
	"vendor",
	"vendor-bin",
	"cache",
	".cache",
	".yarn",
}

type rootCmd struct {
	*cobra.Command

	cfgName     string
	ignoredDirs []string
}

func newRoot() *rootCmd {
	const shortDesc = "Visualize Dependencies and generate deployment orders"

	r := rootCmd{
		Command: &cobra.Command{
			Use:          "dependencies-tool COMMAND",
			Short:        shortDesc,
			SilenceUsage: true,
			Version:      version,
		},
	}

	r.PersistentFlags().StringVar(
		&r.cfgName, "cfg-name",
		filepath.Join("deploy", "deps.yaml"),
		"name or path suffix of the files that are discovered and parsed",
	)
	r.PersistentFlags().StringSliceVar(
		&r.ignoredDirs, "exclude",
		defaultExcludeDirs,
		"comma-separated list of directory names that are excluded when searching for configuration files",
	)
	r.AddCommand(newOrderCmd(&r).Command)
	r.AddCommand(newVerify(&r).Command)
	r.AddCommand(newExportCmd(&r).Command)

	return &r
}

func Execute() {
	cmd := newRoot()
	cmd.SetOut(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
