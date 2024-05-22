package deps

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/simplesurance/dependencies-tool/v3/internal/cfg"
	"github.com/simplesurance/dependencies-tool/v3/internal/datastructs"
	"github.com/simplesurance/dependencies-tool/v3/internal/fs"
	"github.com/simplesurance/dependencies-tool/v3/internal/graphs"
)

// rootVertexName is the start vertex in the graph. The name must be rare to
// prevent that an app exist with the same name.
const rootVertexName string = "root-9e4ecaef-60a4-4300-b0a4-ff3bd1dd7a71"

type Composition struct {
	//Distribution is map of: map[DISTRIBUTION-NAME]:map[APP-NAME]:Dependencies
	Distribution map[string]map[string]*Dependencies `json:"distribution"`
}

// NewComposition creates an empty Composition.
func NewComposition() *Composition {
	return &Composition{Distribution: map[string]map[string]*Dependencies{}}
}

// CompositionFromDir returns a new composition, containing all dependency
// definitions that are found in rootdir or any of it's sub directories.
// Compositions are load from files that match the relative path cfgPath.
// Files that are in directories named as an element in ignoredDirs are ignored.
// CompositionFromDir calls *Composition.Verify() before it returns.
func CompositionFromDir(rootdir string, cfgPath string, ignoredDirs []string) (*Composition, error) {
	realRoot, err := filepath.EvalSymlinks(rootdir)
	if err != nil {
		return nil, err
	}

	cfgPaths, err := fs.Find(realRoot, cfgPath, ignoredDirs)
	if err != nil {
		return nil, fmt.Errorf("discovering %s files failed: %w", cfgPath, err)
	}

	if len(cfgPaths) == 0 {
		return nil, fmt.Errorf("could not find any files in %s matching %s", realRoot, cfgPath)
	}

	comp := NewComposition()
	for _, p := range cfgPaths {
		config, err := cfg.FromFile(p)
		if err != nil {
			return nil, fmt.Errorf("loading config file %q failed: %w", p, err)
		}

		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}

		for distr, deps := range config.Dependencies {
			app, err := dependenciesFromCfg(deps)
			if err != nil {
				return nil, fmt.Errorf("%q: %q: %w", p, distr, err)
			}
			comp.Add(distr, config.AppName, app)
		}
	}

	if err := comp.Verify(); err != nil {
		return nil, err
	}

	return comp, nil
}

// CompositionFromJSON loads a composition from the JSON file filePath.
// Afterwards it calls Composition.Verify.
func CompositionFromJSON(filePath string) (*Composition, error) {
	var comp Composition

	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(fd).Decode(&comp)
	if err != nil {
		return nil, err
	}

	if err := comp.Verify(); err != nil {
		return nil, err
	}

	return &comp, nil
}

// Verify ensures that every soft- and hard dependency is also defined as app
// for a distribution.
func (c *Composition) Verify() error {
	var errs []error

	for distr, apps := range c.Distribution {
		for app, deps := range apps {
			if app == rootVertexName {
				return fmt.Errorf("%q is not allowed as application name", rootVertexName)
			}

			for _, dep := range deps.SoftDeps {
				if _, exist := c.Distribution[distr][dep]; !exist {
					errs = append(errs, fmt.Errorf("%s defines %q as soft dependency for the distribution %q, but %q does not exist or has no %q distribution entry", app, dep, distr, dep, distr))
				}
			}
			for _, dep := range deps.HardDeps {
				if _, exist := c.Distribution[distr][dep]; !exist {
					errs = append(errs, fmt.Errorf("%s defines %q as hard dependency for the distribution %q, but %q does not exist or has no %q distribution entry", app, dep, distr, dep, distr))
				}
			}

		}
	}

	return errors.Join(errs...)
}

func (c *Composition) Add(distribution, appName string, app *Dependencies) {
	distr := c.Distribution[distribution]
	if distr == nil {
		distr = map[string]*Dependencies{}
		c.Distribution[distribution] = distr
	}

	if app == nil {
		app = &Dependencies{}
	}
	distr[appName] = app
}

func (c *Composition) ToJSONFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}

	err = c.ToJSON(f)
	if err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}
func (c *Composition) ToJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(c)
}

// createGraph returns a DAG of the dependencies for the given distribution.
// If apps is empty, the DAG will contain all apps of the distribution.
// Otherwise the DAG will only contain those apps and their dependencies.
// If an element of apps is not found in the distribution an error is returned.
func (c *Composition) createGraph(distribution string, apps []string) (*graphs.Graph, error) {
	distrDeps := c.Distribution[distribution]
	if distrDeps == nil {
		return nil, errors.New("no apps are defined for the distribution")
	}

	g := graphs.NewDigraph()

	// add a parent vertex to the graph, all apps and soft-deps will be
	// connected to it
	g.AddVertex(rootVertexName)

	err := c.forEach(distribution, apps,
		func(appName string, deps *Dependencies) error {
			g.AddVertex(appName)
			g.AddEdge(rootVertexName, appName)
			for _, hd := range deps.HardDeps {
				g.AddVertex(hd)
				g.AddEdge(appName, hd)
			}
			for _, sd := range deps.SoftDeps {
				g.AddVertex(sd)
				g.AddEdge(rootVertexName, sd)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	return g, nil
}

// DependencyOrder calculates the dependency order (reverse topological order)
// for the given distribution and returns it as string slice.
// The dependencies of an app, are ordered before the apps that depend on them.
// If a loop exist between hard dependencies an error is returned.
// If apps is not empty, the order is only calculated for the given app names
// instead of all.
// If an app name is not part of the distribution and error is returned.
func (c *Composition) DependencyOrder(distribution string, apps ...string) ([]string, error) {
	g, err := c.createGraph(distribution, apps)
	if err != nil {
		return nil, err
	}

	sorted, _, err := graphs.TopologicalSort(g)
	if err != nil {
		return nil, err
	}

	order := datastructs.ListToSlice(sorted)

	// remove the rootVertexName start vertex from the list:
	if order[0] != rootVertexName {
		panic(fmt.Sprintf("BUG: first element in the topological sort list is %s, expecting %s\n", order[0], rootVertexName))
	}
	order = order[1:]

	// in Topological order the parent vertex come first, we need the
	// reverse order, dependencies/childs must be ordered before their
	// parents:
	slices.Reverse(order)

	return (order), nil
}

// DependencyOrder calculates the dependency order (reverse topological order)
// for the given distribution and returns it as graph in the dot format.
// Soft dependency are marked with dotted edges.
// If apps is not empty, the order is only calculated for the given app names
// instead of all.
// If an app name is not part of the distribution and error is returned.
// Different from DependencyOrder, not error is returned if a loop exist between
// hard dependencies, the loop shows up in the dot graph.
func (c *Composition) DependencyOrderDot(distribution string, apps ...string) (string, error) {
	graph := graphs.NewDotDiGraph()

	err := c.forEach(distribution, apps,
		func(appName string, deps *Dependencies) error {
			if err := graph.AddNode(appName); err != nil {
				return fmt.Errorf("could not add node %v to graph: %w", appName, err)
			}

			for _, hd := range deps.HardDeps {
				if err := graph.AddNode(hd); err != nil {
					return fmt.Errorf("could not add node %v to graph: %w", hd, err)
				}

				if err := graph.AddEdge(appName, hd); err != nil {
					return fmt.Errorf("could not add edge for hard dependency from %v to %v: %w", appName, hd, err)
				}
			}

			for _, sd := range deps.SoftDeps {
				if err := graph.AddNode(sd); err != nil {
					return fmt.Errorf("could not add node %v to graph: %w", sd, err)
				}

				if err := graph.AddDottedEdge(appName, sd); err != nil {
					return fmt.Errorf("could not add edge for soft dependency from %v to %v: %w", appName, sd, err)
				}
			}

			return nil
		})

	if err != nil {
		return "", err
	}

	return graph.String(), nil
}

func (c *Composition) IsEmpty() bool {
	return len(c.Distribution) == 0
}

// forEach iterates over the applications for the given distribution.
// For each app fn() is called. If fn returns an error, the iteration is
// aborted and the error is returned by forEach.
// If apps is empty or nil, it iterates over all apps of the distribution.
// If apps not empty, it only calls fn for the given app names and their
// recursive dependencies.
// If one of the app names is not found in the distribution an error is
// returned.
func (c *Composition) forEach(distribution string, apps []string, fn func(appName string, deps *Dependencies) error) error {
	if len(apps) > 0 {
		return c.forEachRecursive(distribution, apps, fn)
	}

	distrDeps := c.Distribution[distribution]
	if distrDeps == nil {
		return errors.New("no apps are defined for the distribution")
	}

	for appName, deps := range distrDeps {
		if err := fn(appName, deps); err != nil {
			return err
		}
	}

	return nil
}

func (c *Composition) forEachRecursive(distribution string, apps []string, fn func(appName string, deps *Dependencies) error) error {
	distrDeps := c.Distribution[distribution]
	if distrDeps == nil {
		return errors.New("no apps are defined for the distribution")
	}

	wanted := datastructs.SliceToSet(apps)
	seen := make(map[string]struct{}, len(wanted))
	for len(wanted) > 0 {
		for appName := range wanted {
			deps, exist := distrDeps[appName]
			if !exist {
				return fmt.Errorf("the app does not exist: %s", appName)
			}
			seen[appName] = struct{}{}

			for _, dep := range deps.HardDeps {
				if _, exists := seen[dep]; exists {
					continue
				}
				wanted[dep] = struct{}{}
			}
			for _, dep := range deps.SoftDeps {
				if _, exists := seen[dep]; exists {
					continue
				}
				wanted[dep] = struct{}{}
			}

			if err := fn(appName, deps); err != nil {
				return err
			}
			delete(wanted, appName)
		}
	}

	return nil

}
