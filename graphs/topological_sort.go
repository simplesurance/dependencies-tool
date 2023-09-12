package graphs

import (
	"container/list"
	"errors"
)

var ErrNoDAG = errors.New("graphs: graph is not a DAG")

func TopologicalSort(g *Graph) (topologicalOrder *list.List, topologicalClasses map[string]int, err error) {
	inEdges := make(map[string]int)
	for e := range g.EdgesIter() {
		if _, ok := inEdges[e.Start]; !ok {
			inEdges[e.Start] = 0
		}

		inEdges[e.End]++
	}

	removeEdgesFromVertex := func(v string) {
		for outEdge := range g.HalfedgesIter(v) {
			neighbor := outEdge.End
			inEdges[neighbor]--
		}
	}

	sortedInEdges := sortedEdges(inEdges)

	topologicalClasses = make(map[string]int)
	topologicalOrder = list.New()
	tClass := 0
	for len(inEdges) > 0 {
		topClass := []string{}
		for _, v := range sortedInEdges {
			if _, exist := inEdges[v]; !exist {
				// skip elements that were already processed and
				// deleted from the map, this is expensive but
				// sufficient for our humble usecase
				continue
			}

			inDegree := inEdges[v]
			if inDegree == 0 {
				topClass = append(topClass, v)
				topologicalClasses[v] = tClass
			}
		}
		if len(topClass) == 0 {
			err = ErrNoDAG
			topologicalClasses = make(map[string]int)
			topologicalOrder = list.New()
			return
		}
		for _, v := range topClass {
			removeEdgesFromVertex(v)
			delete(inEdges, v)
			topologicalOrder.PushBack(v)
		}
		tClass++
	}

	return
}
