package kb

import (
	"math"
	"slices"
	"testing"
)

func sampleGraph() []KB {
	must := func(id string, tags ...string) KB {
		kb, err := NewKB(id, "e", "n", "r", tags...)
		if err != nil {
			panic(err)
		}
		return kb
	}
	return []KB{
		must("KB-0001", "calcul", "kpi"),
		must("KB-0002", "calcul"),
		must("KB-0003"),           // isolée : aucun tag
		must("KB-0004", "kpi"),    // liée à 0001 uniquement
		must("KB-0005", "calcul"), // liée à 0001 et 0002
	}
}

func TestRelatedIsSymmetric(t *testing.T) {
	kbs := sampleGraph()
	byID := func(kbs []KB, id string) KB {
		for _, k := range kbs {
			if k.ID == id {
				return k
			}
		}
		t.Fatalf("KB %q absente", id)
		return KB{}
	}
	a := byID(kbs, "KB-0001")
	b := byID(kbs, "KB-0002")

	relA := IDs(Related(kbs, a))
	relB := IDs(Related(kbs, b))

	if !slices.Contains(relA, "KB-0002") || !slices.Contains(relB, "KB-0001") {
		t.Errorf("relation non symétrique : A→%v, B→%v", relA, relB)
	}
}

func TestRelatedExcludesSelfAndUntagged(t *testing.T) {
	kbs := sampleGraph()
	a := kbs[0]
	c := kbs[2]

	if rel := IDs(Related(kbs, a)); slices.Contains(rel, "KB-0001") {
		t.Errorf("Related() inclut la KB elle-même : %v", rel)
	}
	if rel := IDs(Related(kbs, c)); len(rel) != 0 {
		t.Errorf("Related() pour une KB sans tag = %v, want vide", rel)
	}
}

func TestTagStats(t *testing.T) {
	kbs := sampleGraph()

	stats := TagStats(kbs)

	want := []TagStat{{Name: "calcul", Count: 3}, {Name: "kpi", Count: 2}}
	if len(stats) != 2 {
		t.Fatalf("TagStats() = %+v, want 2 tags", stats)
	}
	for i := range want {
		if stats[i] != want[i] {
			t.Errorf("TagStats()[%d] = %+v, want %+v", i, stats[i], want[i])
		}
	}
}

func TestFilterByTag(t *testing.T) {
	kbs := sampleGraph()

	got := IDs(FilterByTag(kbs, "kpi"))

	want := []string{"KB-0001", "KB-0004"}
	if !slices.Equal(got, want) {
		t.Errorf("FilterByTag() = %v, want %v", got, want)
	}
}

func TestGraphNodesOnCircle(t *testing.T) {
	kbs := sampleGraph()[:4] // 4 nœuds triés par ID

	nodes := GraphNodes(kbs)

	if len(nodes) != 4 {
		t.Fatalf("GraphNodes() = %d nœuds, want 4", len(nodes))
	}
	// 1er nœud : angle 0 → (cx + rayon, cy)
	want := GraphNode{ID: "KB-0001", X: graphCenterX + graphRadius, Y: graphCenterY}
	if nodes[0] != want {
		t.Errorf("GraphNodes()[0] = %+v, want %+v", nodes[0], want)
	}
	// Tous les nœuds sont à ~graphRadius du centre
	for _, n := range nodes {
		d := math.Hypot(n.X-graphCenterX, n.Y-graphCenterY)
		if math.Abs(d-graphRadius) > 1e-6 {
			t.Errorf("nœud %s hors cercle : distance %.2f, want %.2f", n.ID, d, graphRadius)
		}
	}
}

func TestGraphEdgesSharedTagsOnly(t *testing.T) {
	kbs := sampleGraph()

	edges := GraphEdges(kbs)

	found := false
	for _, e := range edges {
		if (e.A == "KB-0001" && e.B == "KB-0002") || (e.A == "KB-0002" && e.B == "KB-0001") {
			found = true
		}
	}
	if !found {
		t.Errorf("arête manquante entre KB-0001 et KB-0002 (tag 'calcul' partagé), edges = %+v", edges)
	}

	// Pas d'arête avec la KB sans tag
	for _, e := range edges {
		if e.A == "KB-0003" || e.B == "KB-0003" {
			t.Errorf("arête inattendue avec KB sans tag : %+v", e)
		}
	}
}

func TestTagStatsEmpty(t *testing.T) {
	if stats := TagStats(nil); len(stats) != 0 {
		t.Errorf("TagStats(nil) = %+v, want vide", stats)
	}
}
