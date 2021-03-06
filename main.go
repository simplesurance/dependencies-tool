package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	owndb        bool
	verify       bool
	verifyIgnore string
	deps         string
	sisuDir      string
	compGraph    string
	composeFile  string
	appdirFile   string
	format       string
	environment  string
	region       string
)

func sanitize(in string) string {
	return "\"" + in + "\""
}

func validateParams() error {
	if len(flag.Args()) != 0 {
		return fmt.Errorf("extranous commandline arguments: '%s'", strings.Join(flag.Args(), " "))
	}

	if sisuDir != "" && composeFile != "" {
		return fmt.Errorf("You can only define one of docker-compose and sisu directory")
	}

	if sisuDir != "" && appdirFile != "" {
		return fmt.Errorf("You can only define one of sisu directory and appdir file")
	}

	if composeFile != "" && appdirFile != "" {
		return fmt.Errorf("You can only define one of appdir file and docker-compose file")
	}

	if sisuDir == "" && composeFile == "" && appdirFile == "" {
		return fmt.Errorf("You need to define one of docker-compose, sisu directory and appdir file")
	}

	if format != "text" && format != "dot" {
		return fmt.Errorf("You can only define 'text' or 'dot' as output format")
	}

	return nil
}

func main() {

	flag.BoolVar(&owndb, "owndb", false, "build graph with postgres-db per service (default false)")
	flag.BoolVar(&verify, "verify", false, "verify defined dependencies")

	flag.StringVar(&verifyIgnore, "verify-ignore", "", "comma separated list of service names ignored during verification")
	flag.StringVar(&sisuDir, "sisu", "", "sisu root directory")
	flag.StringVar(&composeFile, "docker-compose", "", "docker-compose file to read from")
	flag.StringVar(&compGraph, "service", "all", "Dependency graph based on [all|<service-name>]")
	flag.StringVar(&deps, "deps", "", "show dependencies of single service")
	flag.StringVar(&appdirFile, "appdirs", "", "path of file with list of appdirs")
	flag.StringVar(&format, "format", "text", "output format ( text or dot )")
	flag.StringVar(&environment, "environment", "", "load deps file named '.deps-<environment>.toml'")
	flag.StringVar(&region, "region", "", "include region in deps file name '.deps-<environment>-<region>'")

	flag.Parse()

	var depsfrom Composition
	var composition Composition

	if err := validateParams(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	composition, err := getComposition()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if owndb {
		composition.PrepareForOwnDb()
	}

	if verify {
		if err := composition.VerifyDependencies(verifyIgnore); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if compGraph == "all" {
		depsfrom = composition
	} else {
		deps, err := composition.RecursiveDepsOf(compGraph)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		depsfrom = *deps
	}

	if deps != "" {
		fmt.Println("[", strings.Join(composition.Deps(deps), ","), "]")
		os.Exit(0)
	}

	if format == "text" {
		secondsorted, err := depsfrom.DeploymentOrder()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not generate graph: %v\n", err)
			os.Exit(1)
		}
		for _, i := range secondsorted {
			fmt.Println(i)
		}
	}

	if format == "dot" {
		fmt.Printf("###########\n# dot of %s\n##########\n", compGraph)
		depsgraph, err := outputDotGraph(depsfrom)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(depsgraph)
	}
}
