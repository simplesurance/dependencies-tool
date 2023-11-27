package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// BaurConf represents the baur.conf and needed options
type BaurConf struct {
	Discover AppDiscover
}

// AppDiscover is a section of baur.toml
type AppDiscover struct {
	AppDirs     []string `toml:"application_dirs"`
	SearchDepth int      `toml:"search_depth"`
}

// FindDepTomls searches for dependencies config files in the AppSearchDirs of the
// AppDiscover and returns all of them
func (d BaurConf) FindDepTomls(dir string) (tomls []string, err error) {

	for _, searchDir := range d.Discover.AppDirs {
		depsCfgs, err := findFilesInSubDir(dir+"/"+searchDir, ".deps*.toml", d.Discover.SearchDepth)
		if err != nil {
			return nil, fmt.Errorf("finding dependencies configs failed %w", err)
		}

		tomls = append(tomls, depsCfgs...)
	}

	return tomls, nil
}

func loadBaurToml(dir string) (d BaurConf, err error) {

	if _, err := toml.DecodeFile(dir, &d); err != nil {
		return d, fmt.Errorf("could not load '%s %w", dir, err)
	}
	return d, nil
}

func depsFileSearchList(region string) []string {
	var fileSearchList []string

	if region != "" && environment != "" {
		fileSearchList = append(fileSearchList, ".deps-"+environment+"-"+region+".toml")
	}

	if region != "" {
		fileSearchList = append(fileSearchList, ".deps-"+region+".toml")
	}

	if environment != "" {
		fileSearchList = append(fileSearchList, ".deps-"+environment+".toml")
	}

	fileSearchList = append(fileSearchList, ".deps.toml")

	return fileSearchList
}

// realDepsToml returns found override deps.toml based on
// given environment and / or region
func realDepsToml(dir, region string) (string, error) {
	filelist := depsFileSearchList(region)

	for _, f := range filelist {
		file := filepath.Join(dir, f)

		if _, err := os.Stat(file); err == nil {
			return file, nil
		} else if !os.IsNotExist(err) {
			return file, err
		}
	}

	return "", fmt.Errorf("could not find one of the following dependency files in %s: %s",
		dir, strings.Join(filelist, ", "))
}

// findFilesInSubDir returns all directories that contain filename that are in
// searchDir. The function descends up to maxdepth levels of directories below
// searchDir
func findFilesInSubDir(searchDir, filename string, maxdepth int) ([]string, error) {
	var result []string
	glob := ""

	for i := 0; i <= maxdepth; i++ {

		globPath := path.Join(searchDir, glob, filename)
		matches, err := filepath.Glob(globPath)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			dir := filepath.Dir(m)
			depsToml, err := realDepsToml(dir, region)
			if err != nil {
				return nil, err
			}
			result = append(result, depsToml)
		}

		glob += "*/"
	}

	return result, nil
}

func applicationTomls(dir string) (tomls []string, err error) {
	r, err := loadBaurToml(dir + "/.baur.toml")
	if err != nil {
		return tomls, err
	}
	return r.FindDepTomls(dir)
}

type tomlService struct {
	Name    string   `toml:"name"`
	TalksTo []string `toml:"talks_to"`
}

func compositionFromSisuDir(directory string) (comp Composition, err error) {
	comp = *NewComposition()

	tomls, err := applicationTomls(directory)
	if err != nil {
		return comp, fmt.Errorf("could not get app tomls, %w", err)
	}

	for _, tomlfile := range tomls {
		var t tomlService
		if _, err := toml.DecodeFile(tomlfile, &t); err != nil {
			return comp, fmt.Errorf("could not toml decode %v, %w", tomlfile, err)
		}
		service := NewService()
		if len(t.TalksTo) > 0 {
			for _, depservice := range t.TalksTo {
				service.AddDependency(depservice)
			}
		}
		comp.AddService(t.Name, service)
	}

	return comp, nil
}
