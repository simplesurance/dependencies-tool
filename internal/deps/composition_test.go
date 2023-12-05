package deps

import (
	"slices"
	"strings"
	"testing"
)

func newTestComp() Composition {
	return Composition{Services: map[string]Service{
		"first-service":  NewService(true, "third"),
		"second-service": NewService(true, "first-service", "consul", "third-service", "postgres"),
		"third-service":  NewService(true, "consul", "postgres"),
		"fourth-service": NewService(true, "fifth-service"),
		"fifth-service":  NewService(true),
	}}
}

func makeTestComp() (comp *Composition) {
	aService := NewService(true, "c")
	bService := NewService(true, "a")
	cService := NewService(true)
	dService := NewService(false, "c")

	comp = NewComposition()
	comp.AddService("b", bService)
	comp.AddService("c", cService)
	comp.AddService("a", aService)
	comp.AddService("d", dService)

	return comp
}

func TestVerifyDependencies(t *testing.T) {
	comp := makeTestComp()
	eService := NewService(true, "notDefined")
	comp.AddService("e", eService)

	if err := comp.VerifyDependencies(); err == nil {
		t.Error("expected error in validation with 'notDefined' service")
	}
}

func TestOutputDotGraph(t *testing.T) {
	comp := makeTestComp()
	dot, _ := OutputDotGraph(*comp, true)

	const expected = "a->c"
	if !strings.Contains(dot, expected) {
		t.Errorf("expected dot to contain %q got %q", expected, dot)
	}
	if !strings.Contains(dot, "d->c") {
		t.Errorf("d->c dependency missing in dot format: %s", dot)
	}

	_, _ = OutputDotGraph(*comp, true)
}

func TestAddDependency(t *testing.T) {
	service := NewService(true, "dep1")

	if _, exists := service.DependsOn["dep1"]; !exists {
		t.Errorf("expected to have 'dep1' in Service.DependsOn got '%v'", service.DependsOn["dep1"])
	}
}

func TestDeps(t *testing.T) {
	comp := makeTestComp()

	got := comp.Deps("a")
	if !Equal(got, []string{"c"}) {
		t.Errorf("expected dep of service a to be %v, got %v", "[\"c\"]", got)
	}
}

func TestRecursiveDepsOf(t *testing.T) {
	comp := makeTestComp()
	exp := []string{"c", "a"}

	got, _ := comp.RecursiveDepsOf("a")
	order, _ := got.DeploymentOrder(false)

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
	exp := []string{"c", "a", "b"}
	got, _ := comp.DeploymentOrder(false)

	if !Equal(exp, got) {
		t.Errorf("expected deployment order of '%v', got '%v'", exp, got)
	}
}

func TestSanitize(t *testing.T) {
	s := "mystring"
	got := s
	exp := "mystring"

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

	if !slices.Contains(testslice, "one") {
		t.Error("expected to be 'one' in testslice")
	}

	if slices.Contains(testslice, "three") {
		t.Error("three is not in testslice")
	}
}

func TestDeployOrderWithUndeployables(t *testing.T) {
	/*
		Dependency structure:
		'(#)' means undeployable

		m (#) -> m1
		m1 (#) -> a, b
		a
		b -> c
		c (#) -> d
		d

		Expected full deploy orders:
		  - without undeployable:
		    d
		    a (does not depend on anything, can be on any position in order)
		    c
		    b

		  - with undeployable:
		    a (can be at any position, does not have deps)
		    d (can be at any position, does not have deps)
		    c
		    b
		    m1
		    m
	*/
	comp := NewComposition()
	m := NewService(false, "m1")
	comp.AddService("m", m)
	m1 := NewService(false, "a", "b")
	comp.AddService("m1", m1)
	a := NewService(true)
	comp.AddService("a", a)
	b := NewService(true, "c")
	comp.AddService("b", b)
	c := NewService(false, "d")
	comp.AddService("c", c)
	d := NewService(true)
	comp.AddService("d", d)

	expectedDeployOrderWithoutUndeployable := []string{"d", "a", "c", "b"}
	expectedDeployOrderWithUndeployable := []string{"d", "c", "b", "a", "m1", "m"}

	orderWithoutUndeployable, err := comp.DeploymentOrder(false)
	fatalOnErr(t, err)
	cmpSlice(t, expectedDeployOrderWithoutUndeployable, orderWithoutUndeployable)

	orderWithUndeployable, err := comp.DeploymentOrder(true)
	fatalOnErr(t, err)
	cmpSlice(t, expectedDeployOrderWithUndeployable, orderWithUndeployable)

	comp, err = comp.RecursiveDepsOf("c")
	fatalOnErr(t, err)
	order, err := comp.DeploymentOrder(true)
	fatalOnErr(t, err)
	cmpSlice(t, []string{"d", "c"}, order)

	comp, err = comp.RecursiveDepsOf("c")
	fatalOnErr(t, err)
	order, err = comp.DeploymentOrder(false)
	fatalOnErr(t, err)
	cmpSlice(t, []string{"d"}, order)

}
