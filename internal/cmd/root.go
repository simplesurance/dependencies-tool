package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/simplesurance/dependencies-tool/v3/internal/cmd/fs"
	"github.com/simplesurance/dependencies-tool/v3/internal/deps"

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

	r.AddCommand(newContainsCmd(&r).Command)
	r.AddCommand(newExportCmd(&r).Command)
	r.AddCommand(newOrderCmd(&r).Command)
	r.AddCommand(newVerify(&r).Command)

	return &r
}

func (r *rootCmd) loadComposition(srcType fs.PathType, src string) (*deps.Composition, error) {
	switch srcType {
	case fs.PathTypeDir:
		return deps.CompositionFromDir(src, r.cfgName, r.ignoredDirs)

	case fs.PathTypeFile:
		return deps.CompositionFromJSON(src)

	default:
		panic(fmt.Sprintf("SrcType has unexpected value: %d", srcType))
	}
}

func Execute() {
	cmd := newRoot()
	cmd.SetOut(os.Stdout)
	if err := cmd.Execute(); err != nil {
		var ee *ErrWithExitCode
		if errors.As(err, &ee) {
			os.Exit(ee.exitCode)
		}

		os.Exit(ExitCodeError)
	}

	os.Exit(ExitCodeSuccess)
}
