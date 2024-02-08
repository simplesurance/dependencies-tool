package deps

import (
	"fmt"

	"github.com/simplesurance/dependencies-tool/v2/internal/cfg"
)

type Dependencies struct {
	SoftDeps []string `json:"soft_dependencies"`
	HardDeps []string `json:"hard_dependencies"`
}

// dependenciesFromCfg converts a map value of the config.Dependencies map to an
// Dependencies struct.
func dependenciesFromCfg(cfgDeps map[string]*cfg.Attributes) (*Dependencies, error) {
	var softdeps, harddeps []string
	for dep, attr := range cfgDeps {
		switch attr.Type {
		case cfg.TypeSoftDependency:
			softdeps = append(softdeps, dep)
		case cfg.TypeHardDependency:
			harddeps = append(harddeps, dep)
		default:
			return nil, fmt.Errorf("%q: unsupported dependency type: %q",
				dep, attr.Type)
		}
	}

	return &Dependencies{SoftDeps: softdeps, HardDeps: harddeps}, nil
}
