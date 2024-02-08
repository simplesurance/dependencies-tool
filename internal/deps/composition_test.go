package deps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/simplesurance/dependencies-tool/v3/internal/testutils"
)

func TestDeploymentOrderNoDeps(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "app1", &Dependencies{})
	comp.Add("prd", "app2", &Dependencies{})
	order, err := comp.DependencyOrder("prd")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"app1", "app2"}, order)
}

func TestDeploymentOrder(t *testing.T) {
	/*
		Dependency structure:
		* means soft-dependency
		m -> m1
		m1 -> a, b
		a
		b -> c*
		c -> d
		d

		Expecting:
		m1 before m
		a,b before m1
		d before c

		c and m can be somewhere in the last, only the other conditions
		must be met.
	*/
	comp := NewComposition()
	comp.Add("prd", "m", &Dependencies{HardDeps: []string{"m1"}})
	comp.Add("prd", "m1", &Dependencies{HardDeps: []string{"a", "b"}})
	comp.Add("prd", "a", &Dependencies{})
	comp.Add("prd", "b", &Dependencies{SoftDeps: []string{"c"}})
	comp.Add("prd", "c", &Dependencies{HardDeps: []string{"d"}})
	comp.Add("prd", "d", &Dependencies{})

	t.Run("all", func(t *testing.T) {
		order, err := comp.DependencyOrder("prd")
		require.NoError(t, err)
		t.Log(order)

		testutils.After(t, order, "m", "m1")
		testutils.After(t, order, "m1", "a")
		testutils.After(t, order, "m1", "b")
		testutils.After(t, order, "c", "d")
		require.Len(t, order, 6)
	})

	t.Run("selected_apps", func(t *testing.T) {
		order, err := comp.DependencyOrder("prd", "m1", "b")
		require.NoError(t, err)
		t.Log(order)
		testutils.After(t, order, "m1", "a")
		testutils.After(t, order, "m1", "b")
		testutils.After(t, order, "c", "d")
		require.Len(t, order, 5)

	})
}

func TestHardDepLoopNotAllowed(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "m", &Dependencies{HardDeps: []string{"m1"}})
	comp.Add("prd", "m1", &Dependencies{HardDeps: []string{"m"}})
	_, err := comp.DependencyOrder("prd")
	t.Log(err)
	require.Error(t, err)
}

func TestSofDepLoopAllowed(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "m", &Dependencies{SoftDeps: []string{"m1"}})
	comp.Add("prd", "m1", &Dependencies{SoftDeps: []string{"m"}})
	order, err := comp.DependencyOrder("prd")
	require.NoError(t, err)
	assert.Contains(t, order, "m")
	assert.Contains(t, order, "m1")
	assert.Len(t, order, 2)
}

func TestVerifyFailsIfSoftDependencyDoesNotExist(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "m", &Dependencies{SoftDeps: []string{"m1"}})
	err := comp.Verify()
	require.Error(t, err)
}

func TestVerifyFailsIfHardDependencyDoesNotExist(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "m", &Dependencies{HardDeps: []string{"m1"}})
	err := comp.Verify()
	require.Error(t, err)
}

func TestOutputDotGraph(t *testing.T) {
	comp := NewComposition()
	comp.Add("prd", "a", &Dependencies{HardDeps: []string{"b"}})
	comp.Add("prd", "b", &Dependencies{HardDeps: []string{"c"}})
	comp.Add("prd", "c", nil)

	t.Run("all", func(t *testing.T) {
		dot, err := comp.DependencyOrderDot("prd")
		require.NoError(t, err)

		for _, expected := range []string{"a->b", "b->c"} {
			assert.Contains(t, dot, expected)
		}
	})

	t.Run("selected_apps", func(t *testing.T) {
		dot, err := comp.DependencyOrderDot("prd", "b")
		require.NoError(t, err)

		assert.Contains(t, dot, "b->c")
		assert.NotContains(t, dot, "a->b")
	})

}
