package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/awalterschulze/gographviz"

	"github.com/simplesurance/dependencies-tool/graphs"
)

// DepService ...
type DepService struct {
	Condition string `json:"condition"`
}

// NewDepService creates new DepService
func NewDepService() DepService {
	return DepService{}
}

// Service ...
type Service struct {
	DependsOn map[string]DepService `json:"depends_on"`
}

// AddDependency adds a service
func (s *Service) AddDependency(name string, service DepService) {
	if _, ok := s.DependsOn[name]; !ok {
		s.DependsOn[name] = service
	}
}

// NewService creates a new Service
func NewService() Service {
	deps := make(map[string]DepService)
	return Service{DependsOn: deps}
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

// PrepareForOwnDb ... clear composition from consul and postgres Service
// Depservices called postgres will be renamed to <servicename>-db
func (comp *Composition) PrepareForOwnDb() {

	for serviceName := range comp.Services {
		if serviceName == "postgres" || serviceName == "consul" {
			delete(comp.Services, serviceName)
			continue
		}

		for depservice := range comp.Services[serviceName].DependsOn {
			if depservice == "postgres" {
				ds := comp.Services[serviceName].DependsOn[depservice]
				comp.Services[serviceName].DependsOn[serviceName+"-db"] = ds
				delete(comp.Services[serviceName].DependsOn, "postgres")
			}
			if depservice == "consul" {
				continue
			}
		}
	}
}

// VerifyDependencies checks if all given dependencies are valid
// it takes a comma separated list of service names.
// These dependencies should be ignored which can be handy when you have external managed ones.
func (comp *Composition) VerifyDependencies(verifyIgnore string) (err error) {
	var services, ignored []string

	for _, n := range strings.Split(verifyIgnore, ",") {
		ignored = append(ignored, strings.TrimSpace(n))
	}

	for serviceName := range comp.Services {
		services = append(services, serviceName)
	}

	for serviceName := range comp.Services {
		for depService := range comp.Services[serviceName].DependsOn {
			if stringsliceContain(ignored, depService) {
				continue
			}
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
			nodeps = append(nodeps, sanitize(serviceName))
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
			services = append(services, sanitize(depservice))
		}
	}
	return services
}

// error if removed service is a dependencies of another service which should not be removed
func removeNotWanted(comp Composition, s string) (todo map[string]bool, err error) {
	todo = make(map[string]bool)
	var notwanted []string
	for _, n := range strings.Split(s, ",") {
		notwanted = append(notwanted, strings.TrimSpace(n))
	}

	for serviceName := range comp.Services {
		if stringsliceContain(notwanted, serviceName) {
			continue
		}
		for depService := range comp.Services[serviceName].DependsOn {
			if stringsliceContain(notwanted, depService) {
				return todo, fmt.Errorf("%s is dependent on %s but not in 'not:' filter list", serviceName, depService)
			}
		}
		todo[serviceName] = true
	}
	return todo, err
}

// RecursiveDepsOf returns Composition with services and dependencies of given servicename
// servicename can be a comma separated list of servicenames
func (comp Composition) RecursiveDepsOf(s string) (newcomp *Composition, err error) {
	var added []string
	todo := make(map[string]bool)

	for _, n := range strings.Split(s, ",") {
		if strings.HasPrefix(n, "not:") {
			todo, err = removeNotWanted(comp, s[4:])
			if err != nil {
				return newcomp, err
			}
			added = append(added, strings.Split(s[4:], ",")...)
			break
		}

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

func sortableGraph(comp Composition) (graph *graphs.Graph, err error) {
	graph = graphs.NewDigraph()

	for service, dependencies := range comp.Services {
		s := sanitize(service)
		graph.AddVertex(s)

		for depservice := range dependencies.DependsOn {
			d := sanitize(depservice)
			graph.AddEdge(s, d)
		}
	}

	//fmt.Printf("graph: %#v", graph)
	return graph, nil
}

func outputDotGraph(comp Composition) (s string, err error) {
	graph := gographviz.NewGraph()
	graph.Name = "G"
	graph.Directed = true

	if err := graph.AddAttr("G", "splines", "\"ortho\""); err != nil {
		return s, fmt.Errorf("could not add Attribute splines: %v", err)
	}
	if err := graph.AddAttr("G", "ranksep", "\"2.0\""); err != nil {
		return s, fmt.Errorf("could not add Attribute ranksep: %v", err)
	}

	for service, dependencies := range comp.Services {
		s := sanitize(service)

		if err := graph.AddNode("G", s, nil); err != nil {
			return s, fmt.Errorf("could not add service %v to graph: %v", service, err)
		}

		for depservice := range dependencies.DependsOn {
			d := sanitize(depservice)

			if !graph.IsNode(d) {
				if err := graph.AddNode("G", d, nil); err != nil {
					return s, fmt.Errorf("could not add service %v to graph: %v", service, err)
				}
			}

			if err := graph.AddEdge(s, d, true, nil); err != nil {
				return s, fmt.Errorf("could not add edge from %v to %v: %v", service, depservice, err)
			}
		}
	}

	return graph.String(), nil
}

func compositionFromDockerComposeOutput(file string) (comp Composition, err error) {
	byteValue, err := os.ReadFile(file)
	if err != nil {
		return comp, fmt.Errorf("could not read file %v", err)
	}

	if err = json.Unmarshal(byteValue, &comp); err != nil {
		return comp, fmt.Errorf("could not unmarshal %v", err)
	}
	return comp, nil
}

func getComposition() (comp Composition, err error) {

	if composeFile != "" {
		return compositionFromDockerComposeOutput(composeFile)
	}

	if sisuDir != "" {
		return compositionFromSisuDir(sisuDir)
	}

	if appdirFile != "" {
		return compositionFromAppdirFile(appdirFile)
	}

	return comp, fmt.Errorf("Do not know what to %v", "do")
}
