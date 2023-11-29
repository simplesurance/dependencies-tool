package deps

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/awalterschulze/gographviz"

	"github.com/simplesurance/dependencies-tool/graphs"
)

// Service ...
type Service struct {
	DependsOn map[string]struct{} `json:"depends_on"`
}

// AddDependency adds a service
func (s *Service) AddDependency(name string) {
	if _, ok := s.DependsOn[name]; !ok {
		s.DependsOn[name] = struct{}{}
	}
}

// NewService creates a new Service
func NewService(deps ...string) Service {
	m := map[string]struct{}{}
	for _, dep := range deps {
		m[dep] = struct{}{}
	}
	return Service{DependsOn: m}
}

// Composition ...
type Composition struct {
	Services map[string]Service `json:"services"`
}

// NewComposition creates a new Composition
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

		service := NewService(t.TalksTo...)
		comp.AddService(t.Name, service)
	}

	return comp, nil
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
			if !stringsliceContain(services, depService) {
				return fmt.Errorf("The app %q was specified as dependency of %q but the app was not found in the repository",
					depService, serviceName,
				)
			}
		}
	}

	return nil
}

// AddService adds a Service
func (comp *Composition) AddService(name string, service Service) {
	if _, ok := comp.Services[name]; !ok {
		comp.Services[name] = service
	}
}

// DeploymentOrder ... deploy from order[0] to order[len(order) -1] :)
func (comp Composition) DeploymentOrder() (order []string, err error) {

	var reverse []string
	var nodeps []string

	for serviceName := range comp.Services {
		if len(comp.Services[serviceName].DependsOn) == 0 {
			nodeps = append(nodeps, serviceName)
		}
	}

	graph, err := sortableGraph(comp)
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
		if !stringsliceContain(reverse, n) {
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

	finished := false

	newcomp = NewComposition()

	for !finished {

		for serviceName := range todo {
			_, ok := comp.Services[serviceName]

			if !ok {
				//return newcomp, fmt.Errorf("Service %v is unknown", serviceName)
				comp.AddService(serviceName, NewService())
			}

			newcomp.Services[serviceName] = comp.Services[serviceName]
			if !stringsliceContain(added, serviceName) {
				added = append(added, serviceName)
			}
			delete(todo, serviceName)

			for name := range comp.Services[serviceName].DependsOn {
				if !stringsliceContain(added, name) {
					todo[name] = true
				}
			}
		}

		if len(todo) == 0 {
			finished = true
		}
	}

	return newcomp, nil
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

func sortableGraph(comp Composition) (graph *graphs.Graph, err error) {
	graph = graphs.NewDigraph()

	for service, dependencies := range comp.Services {
		s := service
		graph.AddVertex(s)

		for depservice := range dependencies.DependsOn {
			d := depservice
			graph.AddEdge(s, d)
		}
	}

	//fmt.Printf("graph: %#v", graph)
	return graph, nil
}

func OutputDotGraph(comp Composition) (s string, err error) {
	graph := gographviz.NewGraph()
	graph.Name = "G"
	graph.Directed = true

	if err := graph.AddAttr("G", "splines", "\"ortho\""); err != nil {
		return s, fmt.Errorf("could not add Attribute splines: %w", err)
	}
	if err := graph.AddAttr("G", "ranksep", "\"2.0\""); err != nil {
		return s, fmt.Errorf("could not add Attribute ranksep: %w", err)
	}

	for service, dependencies := range comp.Services {
		s := service

		if err := graph.AddNode("G", s, nil); err != nil {
			return s, fmt.Errorf("could not add service %v to graph: %w", service, err)
		}

		for depservice := range dependencies.DependsOn {
			d := depservice

			if !graph.IsNode(d) {
				if err := graph.AddNode("G", d, nil); err != nil {
					return s, fmt.Errorf("could not add service %v to graph: %w", service, err)
				}
			}

			if err := graph.AddEdge(s, d, true, nil); err != nil {
				return s, fmt.Errorf("could not add edge from %v to %v: %w", service, depservice, err)
			}
		}
	}

	return graph.String(), nil
}
