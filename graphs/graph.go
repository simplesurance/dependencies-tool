package graphs

import (
	"fmt"
	"slices"
)

// An Edge connects two vertices.
type Edge struct {
	Start string
	End   string
}

// A Halfedge is an edge where just the end vertex is
// stored. The start vertex is inferred from the context.
type Halfedge struct {
	End string
}

// A Graph is defined by its vertices and edges stored as
// an adjacency set.
type Graph struct {
	Adjacency map[string]*Set
	Directed  bool
}

// NewGraph creates a new empty graph.
func NewGraph() *Graph {
	return &Graph{
		Adjacency: map[string]*Set{},
		Directed:  false,
	}
}

// NewDigraph creates a new empty directed graph.
func NewDigraph() *Graph {
	graph := NewGraph()
	graph.Directed = true
	return graph
}

// AddVertex adds the given vertex to the graph.
func (g *Graph) AddVertex(v string) {
	if _, exists := g.Adjacency[v]; !exists {
		g.Adjacency[v] = NewSet()
	}
}

// AddEdge adds an edge to the graph. The edge connects
// vertex v1 and vertex v2.
func (g *Graph) AddEdge(v1, v2 string) {
	g.AddVertex(v1)
	g.AddVertex(v2)

	g.Adjacency[v1].Add(Halfedge{
		End: v2,
	})

	if !g.Directed {
		g.Adjacency[v2].Add(Halfedge{
			End: v1,
		})
	}
}

// Dump prints all edges with to stdout.
func (g *Graph) Dump() {
	for e := range g.EdgesIter() {
		fmt.Printf("(%v,%v)\n", e.Start, e.End)
	}
}

// NVertices returns the number of vertices.
func (g *Graph) NVertices() int {
	return len(g.Adjacency)
}

// NEdges returns the number of edges.
func (g *Graph) NEdges() int {
	n := 0

	for _, v := range g.Adjacency {
		n += v.Len()
	}

	// Don’t count a-b and b-a edges for undirected graphs
	// as two separate edges.
	if !g.Directed {
		n /= 2
	}

	return n
}

func sortedEdges(m map[string]int) []string {
	sortedStrs := make([]string, 0, len(m))

	for k := range m {
		sortedStrs = append(sortedStrs, k)
	}

	slices.Sort(sortedStrs)
	return sortedStrs
}

// EdgesIter returns a channel with all edges of the graph.
func (g *Graph) EdgesIter() chan Edge {
	ch := make(chan Edge)
	go func() {
		for v, s := range g.Adjacency {
			for x := range s.Iter() {
				he := x.(Halfedge)
				ch <- Edge{v, he.End}
			}
		}
		close(ch)
	}()
	return ch
}

// HalfedgesIter returns a channel with all halfedges for
// the given start vertex.
func (g *Graph) HalfedgesIter(v string) chan Halfedge {
	ch := make(chan Halfedge)
	go func() {
		if s, exists := g.Adjacency[v]; exists {
			for x := range s.Iter() {
				he := x.(Halfedge)
				ch <- he
			}
		}
		close(ch)
	}()
	return ch
}
