package graphs

import (
	"github.com/awalterschulze/gographviz"
)

type Dot struct {
	g         *gographviz.Escape
	graphName string
}

func NewDotDiGraph() *Dot {
	const graphName = "G"
	g := gographviz.NewEscape()
	g.Directed = true
	_ = g.SetName(graphName) // SetName never returns an error

	_ = g.AddAttr(graphName, "splines", "ortho")
	_ = g.AddAttr(graphName, "ranksep", "2.0")

	return &Dot{
		g:         g,
		graphName: graphName,
	}
}

func (g *Dot) AddNode(name string) error {
	if g.g.IsNode(name) {
		return nil
	}
	return g.g.AddNode(g.graphName, name, nil)
}

func (g *Dot) AddEdge(src, dest string) error {
	return g.g.AddEdge(src, dest, true, nil)
}

func (g *Dot) AddDottedEdge(src, dest string) error {
	return g.g.AddEdge(src, dest, true, map[string]string{"style": "dotted"})
}

func (g *Dot) String() string {
	return g.g.String()
}
