package deps

import (
	"strings"
	"testing"
)

func newTestComp() Composition {
	return Composition{Services: map[string]Service{
		"first-service":  NewService("third"),
		"second-service": NewService("first-service", "consul", "third-service", "postgres"),
		"third-service":  NewService("consul", "postgres"),
		"fourth-service": NewService("fifth-service"),
		"fifth-service":  NewService(),
	}}
}

func makeTestComp() (comp *Composition) {
	aService := NewService("c")
	bService := NewService("a")
	cService := NewService()

	comp = NewComposition()
	comp.AddService("b", bService)
	comp.AddService("c", cService)
	comp.AddService("a", aService)

	return comp
}

func TestVerifyDependencies(t *testing.T) {
	comp := makeTestComp()
	dService := NewService("notDefined")
	comp.AddService("d", dService)

	if err := comp.VerifyDependencies(); err == nil {
		t.Error("expected error in validation with 'notDefined' service")
	}
}

// //
func TestOutputDotGraph(t *testing.T) {
	comp := makeTestComp()
	dot, _ := OutputDotGraph(*comp)

	if !strings.Contains(dot, "\"a\"->\"c\"") {
		t.Errorf("expected dot to contain '\"a\"->\"c\"' got %v", dot)
	}
	_, _ = OutputDotGraph(*comp)
}

func TestAddDependency(t *testing.T) {
	service := NewService("dep1")

	if _, exists := service.DependsOn["dep1"]; !exists {
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
	comp := newTestComp()

	got, _ := comp.RecursiveDepsOf("fourth-service,first-service")

	_, ok := got.Services["fifth-service"]
	if !ok {
		t.Error("expected to have 'fifth-service' in composition")
	}
}

func TestRecursiveDepsOfWithListOfServicesAndBlank(t *testing.T) {
	comp := newTestComp()
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

func TestSanitize(t *testing.T) {
	s := "mystring"
	got := sanitize(s)
	exp := "\"mystring\""

	if got != exp {
		t.Errorf("expected to be result %v , got %v", exp, got)
	}
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestIsIn(t *testing.T) {
	testslice := []string{"one", "two"}

	if !stringsliceContain(testslice, "one") {
		t.Error("expected to be 'one' in testslice")
	}

	if stringsliceContain(testslice, "three") {
		t.Error("three is not in testslice")
	}
}
