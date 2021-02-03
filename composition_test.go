// +build !integrationtest unittest

package main

import (
	"strings"
	"testing"
)

func makeTestComp() (comp *Composition) {
	aService := NewService()
	bService := NewService()
	cService := NewService()

	aService.AddDependency("c", NewDepService())
	bService.AddDependency("a", NewDepService())

	comp = NewComposition()
	comp.AddService("b", bService)
	comp.AddService("c", cService)
	comp.AddService("a", aService)

	return comp
}

func TestVerifyDependencies(t *testing.T) {
	comp := makeTestComp()
	dService := NewService()
	dService.AddDependency("notDefined", NewDepService())
	comp.AddService("d", dService)

	if err := comp.VerifyDependencies(""); err == nil {
		t.Error("expected error in validation without ignoring 'notDefined' service")
	}

	if err := comp.VerifyDependencies("notDefined"); err != nil {
		t.Error("expected no error in validation with ignoring 'notDefined' service")
	}
}

////
func TestOutputDotGraph(t *testing.T) {
	comp := makeTestComp()
	dot, _ := outputDotGraph(*comp)

	if !strings.Contains(dot, "\"a\"->\"c\"") {
		t.Errorf("expected dot to contain '\"a\"->\"c\"' got %v", dot)
	}
	// add consul but ignore in dot
	comp.AddService("consul", NewService())
	comp.PrepareForOwnDb()
	dot2, _ := outputDotGraph(*comp)
	if strings.Contains(dot2, "consul") {
		t.Errorf("expected no consul in dotgraph, got %v", dot2)
	}
}

func TestPrepareForOwnDb(t *testing.T) {
	a := NewService()
	a.AddDependency("postgres", NewDepService())
	comp := NewComposition()
	comp.AddService("a", a)
	comp.PrepareForOwnDb()
	dot, _ := outputDotGraph(*comp)
	if !strings.Contains(dot, "a-db") {
		t.Error("expected an own db for service a in dotgraph")
	}
}

func TestAddDependency(t *testing.T) {
	service := NewService()
	dep := NewDepService()

	service.AddDependency("dep1", dep)

	if service.DependsOn["dep1"] != dep {
		t.Errorf("expected to have 'dep1' in Service.DependsOn got '%v'", service.DependsOn["dep1"])
	}
}

func TestDeps(t *testing.T) {
	comp := makeTestComp()

	got := comp.Deps("a")
	if !Equal(got, []string{"\"c\""}) {
		t.Errorf("expected dep of service a to be %v, got %v", "[\"c\"]", got)
	}
}

func TestRecursiveDepsOf(t *testing.T) {

	comp := makeTestComp()
	exp := []string{"\"c\"", "\"a\""}

	got, _ := comp.RecursiveDepsOf("a")
	order, _ := got.DeploymentOrder()

	if !Equal(order, exp) {
		t.Errorf("expected deps of A equal %v, got %v", exp, order)
	}
}

func TestRecursiveDepsOfWithListOfServices(t *testing.T) {
	comp, _ := compositionFromDockerComposeOutput("test/working-compose.json")
	got, _ := comp.RecursiveDepsOf("fourth-service,first-service")

	_, ok := got.Services["fifth-service"]
	if !ok {
		t.Error("expected to have 'fifth-service' in composition")
	}

	_, err := comp.RecursiveDepsOf("not:first-service")
	if err == nil {
		t.Error("expected error with 'not:first-service' ")
	}
}

func TestRecursiveDepsOfWithNot(t *testing.T) {
	comp, _ := compositionFromDockerComposeOutput("test/working-compose.json")

	notGot, _ := comp.RecursiveDepsOf("not:first-service,second-service")

	_, ok := notGot.Services["first-service"]
	if ok {
		t.Error("expected not to have 'first-service' in composition")
	}
}

func TestRecursiveDepsOfWithListOfServicesAndBlank(t *testing.T) {
	comp, _ := compositionFromDockerComposeOutput("test/working-compose.json")
	got, _ := comp.RecursiveDepsOf("fifth-service, fourth-service")

	_, ok := got.Services["fourth-service"]
	if !ok {
		t.Error("expected to have 'fourth-service' in composition")
	}
}

func TestDeployOrder(t *testing.T) {
	comp := makeTestComp()

	exp := []string{"\"c\"", "\"a\"", "\"b\""}
	got, _ := comp.DeploymentOrder()

	if !Equal(exp, got) {
		t.Errorf("expected deployment order of '%v', got '%v'", exp, got)
	}
}

func TestCompositionFromDockerComposeOutput(t *testing.T) {
	_, err := compositionFromDockerComposeOutput("test/working-compose.json")
	if err != nil {
		t.Errorf("expected to get working composition from test/working-compose.json. got %v", err)
	}

	_, err = compositionFromDockerComposeOutput("test/nonexistant")
	if err == nil {
		t.Error("expected to fail from test/nonexistant. ")
	}

	_, err = compositionFromDockerComposeOutput("test/broken.json")
	if err == nil {
		t.Error("expected to fail from test/broken.json. ")
	}
}

func TestGetComposition(t *testing.T) {
	composeFile = "test/working-compose.json"
	_, err := getComposition()
	if err != nil {
		t.Errorf("expected no error with compose file %v, got %v", composeFile, err)
	}
	composeFile = ""

	sisuDir = "test"
	_, err = getComposition()
	if err != nil {
		t.Errorf("expected no error with sisu dir %v, got %v", sisuDir, err)
	}
	sisuDir = ""

	appdirFile = "test/appdirfile"
	_, err = getComposition()
	if err != nil {
		t.Errorf("expected no error with appdirFile %v, got %v", appdirFile, err)
	}
	appdirFile = ""

	_, err = getComposition()
	if err == nil {
		t.Error("This should not happen because of validation")
	}
}

func TestRemoveNotWanted(t *testing.T) {
	comp, _ := compositionFromDockerComposeOutput("test/working-compose.json")
	var list []string

	mapList, _ := removeNotWanted(comp, "first-service,third")
	for k := range mapList {
		list = append(list, k)
	}

	if stringsliceContain(list, "first-service") {
		t.Errorf("expected list to not contain 'first-service'. list: %v", list)
	}
}
