package dag

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/muratcelep/terraform/not-internal/tfdiags"

	_ "github.com/muratcelep/terraform/not-internal/logging"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestAcyclicGraphRoot(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))

	if root, err := g.Root(); err != nil {
		t.Fatalf("err: %s", err)
	} else if root != 3 {
		t.Fatalf("bad: %#v", root)
	}
}

func TestAcyclicGraphRoot_cycle(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(2, 3))
	g.Connect(BasicEdge(3, 1))

	if _, err := g.Root(); err == nil {
		t.Fatal("should error")
	}
}

func TestAcyclicGraphRoot_multiple(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))

	if _, err := g.Root(); err == nil {
		t.Fatal("should error")
	}
}

func TestAyclicGraphTransReduction(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(1, 3))
	g.Connect(BasicEdge(2, 3))
	g.TransitiveReduction()

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testGraphTransReductionStr)
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

func TestAyclicGraphTransReduction_more(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Add(4)
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(1, 3))
	g.Connect(BasicEdge(1, 4))
	g.Connect(BasicEdge(2, 3))
	g.Connect(BasicEdge(2, 4))
	g.Connect(BasicEdge(3, 4))
	g.TransitiveReduction()

	actual := strings.TrimSpace(g.String())
	expected := strings.TrimSpace(testGraphTransReductionMoreStr)
	if actual != expected {
		t.Fatalf("bad: %s", actual)
	}
}

// use this to simulate slow sort operations
type counter struct {
	Name  string
	Calls int64
}

func (s *counter) String() string {
	s.Calls++
	return s.Name
}

// Make sure we can reduce a sizable, fully-connected graph.
func TestAyclicGraphTransReduction_fullyConnected(t *testing.T) {
	var g AcyclicGraph

	const nodeCount = 200
	nodes := make([]*counter, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodes[i] = &counter{Name: strconv.Itoa(i)}
	}

	// Add them all to the graph
	for _, n := range nodes {
		g.Add(n)
	}

	// connect them all
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			g.Connect(BasicEdge(nodes[i], nodes[j]))
		}
	}

	g.TransitiveReduction()

	vertexNameCalls := int64(0)
	for _, n := range nodes {
		vertexNameCalls += n.Calls
	}

	switch {
	case vertexNameCalls > 2*nodeCount:
		// Make calling it more the 2x per node fatal.
		// If we were sorting this would give us roughly ln(n)(n^3) calls, or
		// >59000000 calls for 200 vertices.
		t.Fatalf("VertexName called %d times", vertexNameCalls)
	case vertexNameCalls > 0:
		// we don't expect any calls, but a change here isn't necessarily fatal
		t.Logf("WARNING: VertexName called %d times", vertexNameCalls)
	}
}

func TestAcyclicGraphValidate(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))

	if err := g.Validate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAcyclicGraphValidate_cycle(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(2, 1))

	if err := g.Validate(); err == nil {
		t.Fatal("should error")
	}
}

func TestAcyclicGraphValidate_cycleSelf(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Connect(BasicEdge(1, 1))

	if err := g.Validate(); err == nil {
		t.Fatal("should error")
	}
}

func TestAcyclicGraphAncestors(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Add(4)
	g.Add(5)
	g.Connect(BasicEdge(0, 1))
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(2, 3))
	g.Connect(BasicEdge(3, 4))
	g.Connect(BasicEdge(4, 5))

	actual, err := g.Ancestors(2)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}

	expected := []Vertex{3, 4, 5}

	if actual.Len() != len(expected) {
		t.Fatalf("bad length! expected %#v to have len %d", actual, len(expected))
	}

	for _, e := range expected {
		if !actual.Include(e) {
			t.Fatalf("expected: %#v to include: %#v", expected, actual)
		}
	}
}

func TestAcyclicGraphDescendents(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Add(4)
	g.Add(5)
	g.Connect(BasicEdge(0, 1))
	g.Connect(BasicEdge(1, 2))
	g.Connect(BasicEdge(2, 3))
	g.Connect(BasicEdge(3, 4))
	g.Connect(BasicEdge(4, 5))

	actual, err := g.Descendents(2)
	if err != nil {
		t.Fatalf("err: %#v", err)
	}

	expected := []Vertex{0, 1}

	if actual.Len() != len(expected) {
		t.Fatalf("bad length! expected %#v to have len %d", actual, len(expected))
	}

	for _, e := range expected {
		if !actual.Include(e) {
			t.Fatalf("expected: %#v to include: %#v", expected, actual)
		}
	}
}

func TestAcyclicGraphWalk(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(3, 1))

	var visits []Vertex
	var lock sync.Mutex
	err := g.Walk(func(v Vertex) tfdiags.Diagnostics {
		lock.Lock()
		defer lock.Unlock()
		visits = append(visits, v)
		return nil
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := [][]Vertex{
		{1, 2, 3},
		{2, 1, 3},
	}
	for _, e := range expected {
		if reflect.DeepEqual(visits, e) {
			return
		}
	}

	t.Fatalf("bad: %#v", visits)
}

func TestAcyclicGraphWalk_error(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Add(4)
	g.Connect(BasicEdge(4, 3))
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(2, 1))

	var visits []Vertex
	var lock sync.Mutex
	err := g.Walk(func(v Vertex) tfdiags.Diagnostics {
		lock.Lock()
		defer lock.Unlock()

		var diags tfdiags.Diagnostics

		if v == 2 {
			diags = diags.Append(fmt.Errorf("error"))
			return diags
		}

		visits = append(visits, v)
		return diags
	})
	if err == nil {
		t.Fatal("should error")
	}

	expected := []Vertex{1}
	if !reflect.DeepEqual(visits, expected) {
		t.Errorf("wrong visits\ngot:  %#v\nwant: %#v", visits, expected)
	}

}

func BenchmarkDAG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		count := 150
		b.StopTimer()
		g := &AcyclicGraph{}

		// create 4 layers of fully connected nodes
		// layer A
		for i := 0; i < count; i++ {
			g.Add(fmt.Sprintf("A%d", i))
		}

		// layer B
		for i := 0; i < count; i++ {
			B := fmt.Sprintf("B%d", i)
			g.Add(B)
			for j := 0; j < count; j++ {
				g.Connect(BasicEdge(B, fmt.Sprintf("A%d", j)))
			}
		}

		// layer C
		for i := 0; i < count; i++ {
			c := fmt.Sprintf("C%d", i)
			g.Add(c)
			for j := 0; j < count; j++ {
				// connect them to previous layers so we have something that requires reduction
				g.Connect(BasicEdge(c, fmt.Sprintf("A%d", j)))
				g.Connect(BasicEdge(c, fmt.Sprintf("B%d", j)))
			}
		}

		// layer D
		for i := 0; i < count; i++ {
			d := fmt.Sprintf("D%d", i)
			g.Add(d)
			for j := 0; j < count; j++ {
				g.Connect(BasicEdge(d, fmt.Sprintf("A%d", j)))
				g.Connect(BasicEdge(d, fmt.Sprintf("B%d", j)))
				g.Connect(BasicEdge(d, fmt.Sprintf("C%d", j)))
			}
		}

		b.StartTimer()
		// Find dependencies for every node
		for _, v := range g.Vertices() {
			_, err := g.Ancestors(v)
			if err != nil {
				b.Fatal(err)
			}
		}

		// reduce the final graph
		g.TransitiveReduction()
	}
}

func TestAcyclicGraph_ReverseDepthFirstWalk_WithRemoval(t *testing.T) {
	var g AcyclicGraph
	g.Add(1)
	g.Add(2)
	g.Add(3)
	g.Connect(BasicEdge(3, 2))
	g.Connect(BasicEdge(2, 1))

	var visits []Vertex
	var lock sync.Mutex
	err := g.SortedReverseDepthFirstWalk([]Vertex{1}, func(v Vertex, d int) error {
		lock.Lock()
		defer lock.Unlock()
		visits = append(visits, v)
		g.Remove(v)
		return nil
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []Vertex{1, 2, 3}
	if !reflect.DeepEqual(visits, expected) {
		t.Fatalf("expected: %#v, got: %#v", expected, visits)
	}
}

const testGraphTransReductionStr = `
1
  2
2
  3
3
`

const testGraphTransReductionMoreStr = `
1
  2
2
  3
3
  4
4
`
