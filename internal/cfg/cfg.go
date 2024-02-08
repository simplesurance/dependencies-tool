package cfg

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const TypeSoftDependency = "soft"
const TypeHardDependency = "hard"
const TypeDefaultDependency = TypeHardDependency

type Attributes struct {
	Type string `yaml:"type,flow"`
}

// Config represents a decoded configuration file, that declares the
// dependencies of an Application.
type Config struct {
	AppName string `yaml:"name"`
	// Key of Dependencise must either be "Default" (case-insensitive) a
	// string existing in Targets
	// Dependencies is map of map[DISTRIBUTION-NAME]map[DEPENDS-ON-APP-NAME]Attributes
	Dependencies map[string]map[string]*Attributes `yaml:"dependencies"`
}

// Unmarshal reads and decodes a YAML marshalled Config struct. marshalled
// Config struct from r.
func Unmarshal(r io.Reader) (_ *Config, err error) {
	var result Config
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)

	defer func() {
		// sadly Decode panics on some formatting issues in the input,
		// catch the panic and return it as an error
		r := recover()
		if r == nil {
			return
		}

		switch tr := r.(type) {
		case error:
			err = tr
		default:
			err = fmt.Errorf("yaml decoding failed: %s", r)
		}
	}()

	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	result.setDefaults()

	return &result, nil
}

// FromFile unmarshals a YAML encoded configuration from the file at path.
func FromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Unmarshal(f)
}

// setDefaults sets the Attributes.Type default value in a to
// TypeDefaultDependency, if the Type field or the pointer to the Attributes
// struct is unset.
func (a *Config) setDefaults() {
	for _, dep := range a.Dependencies {
		for app, attr := range dep {
			if attr == nil {
				dep[app] = &Attributes{Type: TypeHardDependency}
				continue
			}
			if attr.Type == "" {
				attr.Type = TypeDefaultDependency
				continue
			}
		}
	}
}

func (a *Config) Validate() error {
	if strings.TrimSpace(a.AppName) == "" {
		return fmt.Errorf("name is empty or contains only whitespaces: %q", a.AppName)
	}

	if len(a.Dependencies) == 0 {
		return fmt.Errorf("dependencies map is empty, expecting at least 1 distribution key")
	}

	for distr, mApp := range a.Dependencies {
		if strings.TrimSpace(distr) == "" {
			return fmt.Errorf("distribution is empty or contains only whitespaces: %q", distr)
		}

		if len(mApp) == 0 {
			continue
		}

		for depApp, attr := range mApp {
			if strings.TrimSpace(depApp) == "" {
				return fmt.Errorf("dependencies[%s] entry key is empty or contains only whitespaces: %q",
					distr, depApp)
			}

			if attr != nil && attr.Type != TypeHardDependency && attr.Type != TypeSoftDependency {
				return fmt.Errorf("dependencies[%s][%s].type is %q, expecting %q, %q or an null map value",
					distr, depApp, attr.Type, TypeHardDependency, TypeSoftDependency)
			}
		}
	}

	return nil
}
