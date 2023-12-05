package deps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/awalterschulze/gographviz"

	"github.com/simplesurance/dependencies-tool/graphs"
)

type Service struct {
	DependsOn  map[string]struct{} `json:"depends_on"`
	Deployable bool                `json:"deployable"`
}

// AddDependency declares name to be a dependency of s.
func (s *Service) AddDependency(name string) {
	if _, ok := s.DependsOn[name]; !ok {
		s.DependsOn[name] = struct{}{}
	}
}

// NewService creates a Service
func NewService(deployable bool, deps ...string) Service {
	m := map[string]struct{}{}
	for _, dep := range deps {
		m[dep] = struct{}{}
	}
	return Service{Deployable: deployable, DependsOn: m}
}

type Composition struct {
	Services map[string]Service `json:"services"`
}

// NewComposition creates a Composition
func NewComposition() *Composition {
	svs := make(map[string]Service)
	return &Composition{Services: svs}
}

func CompositionFromSisuDir(directory, env, region string) (*Composition, error) {
	tomls, err := applicationTomls(directory, env, region)
	if err != nil {
		return nil, fmt.Errorf("could not get app tomls, %w", err)
	}

	comp := NewComposition()
	for _, tomlfile := range tomls {
		var t tomlService
		if _, err := toml.DecodeFile(tomlfile, &t); err != nil {
			return nil, fmt.Errorf("could not toml decode %v, %w", tomlfile, err)
		}

		isDeployable := dirExists(filepath.Join(filepath.Dir(tomlfile), "deploy"))

		service := NewService(isDeployable, t.TalksTo...)
		comp.AddService(t.Name, service)
	}

	return comp, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func CompositionFromJSON(filePath string) (*Composition, error) {
	var result Composition

	fd, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(fd).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// VerifyDependencies checks if all given dependencies are valid
// it takes a comma separated list of service names.
// These dependencies should be ignored which can be handy when you have external managed ones.
func (comp *Composition) VerifyDependencies() (err error) {
	services := make([]string, 0, len(comp.Services))
	for serviceName := range comp.Services {
		services = append(services, serviceName)
	}

	for serviceName := range comp.Services {
		for depService := range comp.Services[serviceName].DependsOn {
			if !slices.Contains(services, depService) {
				return fmt.Errorf("The app %q was specified as dependency of %q but the app was not found in the repository",
					depService, serviceName,
				)
			}
		}
	}

	return nil
}

// AddService adds a Service with the given name to the Composition.
func (comp *Composition) AddService(name string, service Service) {
	if _, ok := comp.Services[name]; !ok {
		comp.Services[name] = service
	}
}

// DeploymentOrder ... deploy from order[0] to order[len(order) -1] :)
func (comp Composition) DeploymentOrder(includeUndeployable bool) (order []string, err error) {
	var reverse []string
	var nodeps []string

	for serviceName := range comp.Services {
		service := comp.Services[serviceName]
		if len(service.DependsOn) == 0 {
			if includeUndeployable || service.Deployable {
				nodeps = append(nodeps, serviceName)
			}
		}
	}

	graph, err := sortableGraph(comp, includeUndeployable)
	if err != nil {
		return order, err
	}

	topOrder, _, err := graphs.TopologicalSort(graph)
	if err != nil {
		return order, err
	}

	e := topOrder.Front()

	for e != nil {
		reverse = append(reverse, e.Value.(string))
		e = e.Next()
	}

	slices.Sort(nodeps)
	// graphs.TopologicalSort() deletes services which are no dependencies
	// and have no dependencies so we add them again
	for _, n := range nodeps {
		if !slices.Contains(reverse, n) {
			reverse = append(reverse, n)
		}
	}

	for i := len(reverse); i >= 1; i-- {
		order = append(order, reverse[i-1])
	}

	return order, nil
}

// Deps returns list of dependent services
func (comp Composition) Deps(s string) (services []string) {
	if _, ok := comp.Services[s]; ok {
		for depservice := range comp.Services[s].DependsOn {
			services = append(services, depservice)
		}
	}
	return services
}

// RecursiveDepsOf returns Composition with services and dependencies of given servicename
// servicename can be a comma separated list of servicenames
func (comp Composition) RecursiveDepsOf(s string) (newcomp *Composition, err error) {
	var added []string
	todo := make(map[string]bool)

	for _, n := range strings.Split(s, ",") {
		todo[strings.TrimSpace(n)] = true
	}

	newcomp = NewComposition()

	for len(todo) > 0 {
		for serviceName := range todo {
			service, ok := comp.Services[serviceName]

			if !ok {
				comp.AddService(serviceName, NewService(service.Deployable))
			}

			newcomp.Services[serviceName] = comp.Services[serviceName]
			if !slices.Contains(added, serviceName) {
				added = append(added, serviceName)
			}
			delete(todo, serviceName)

			for name := range comp.Services[serviceName].DependsOn {
				if !slices.Contains(added, name) {
					todo[name] = true
				}
			}
		}

	}

	return newcomp, nil
}

func (comp *Composition) isDeployable(serviceName string) bool {
	return comp.Services[serviceName].Deployable
}

func (comp *Composition) ToJSONFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(comp)
	if err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}

func sortableGraph(comp Composition, includeUndeployable bool) (graph *graphs.Graph, err error) {
	graph = graphs.NewDigraph()

	for serviceName, service := range comp.Services {
		if !includeUndeployable && !service.Deployable {
			continue
		}

		graph.AddVertex(serviceName)

		for depservice := range service.DependsOn {
			if includeUndeployable || comp.isDeployable(serviceName) {
				graph.AddEdge(serviceName, depservice)
			}
		}
	}

	return graph, nil
}

func OutputDotGraph(comp Composition, includeUndeployable bool) (s string, err error) {
	graph := gographviz.NewGraph()
	graph.Name = "G"
	graph.Directed = true

	if err := graph.AddAttr("G", "splines", "\"ortho\""); err != nil {
		return s, fmt.Errorf("could not add Attribute splines: %w", err)
	}
	if err := graph.AddAttr("G", "ranksep", "\"2.0\""); err != nil {
		return s, fmt.Errorf("could not add Attribute ranksep: %w", err)
	}

	for serviceName, service := range comp.Services {
		s := serviceName

		if !includeUndeployable && !service.Deployable {
			continue
		}

		if err := graph.AddNode("G", s, nil); err != nil {
			return s, fmt.Errorf("could not add service %v to graph: %w", serviceName, err)
		}

		for depservice := range service.DependsOn {
			d := depservice
			if !includeUndeployable && !service.Deployable {
				continue
			}

			if !graph.IsNode(d) {
				if err := graph.AddNode("G", d, nil); err != nil {
					return s, fmt.Errorf("could not add service %v to graph: %w", serviceName, err)
				}
			}

			if err := graph.AddEdge(s, d, true, nil); err != nil {
				return s, fmt.Errorf("could not add edge from %v to %v: %w", serviceName, depservice, err)
			}
		}
	}

	return graph.String(), nil
}
