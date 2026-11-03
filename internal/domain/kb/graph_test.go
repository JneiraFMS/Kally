package kb

import (
	"fmt"
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

// TestFilterByAllTags : avec plusieurs tags, une KB n'est conservée que si
// elle porte TOUS les tags (requête AND, à la Logseq) ; sans tag, tout est
// retourné.
func TestFilterByAllTags(t *testing.T) {
	kbs := sampleGraph()

	got := IDs(FilterByAllTags(kbs, "calcul", "kpi"))
	want := []string{"KB-0001"}
	if !slices.Equal(got, want) {
		t.Errorf("FilterByAllTags(calcul, kpi) = %v, want %v", got, want)
	}

	// Un seul tag : équivalent à FilterByTag
	if got, want := IDs(FilterByAllTags(kbs, "calcul")), []string{"KB-0001", "KB-0002", "KB-0005"}; !slices.Equal(got, want) {
		t.Errorf("FilterByAllTags(calcul) = %v, want %v", got, want)
	}

	// Sans tag : aucun filtrage
	if got := IDs(FilterByAllTags(kbs)); len(got) != len(kbs) {
		t.Errorf("FilterByAllTags() sans tag = %v, want toutes (%d)", got, len(kbs))
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

func TestRelatedOrdersBySharedTagCountThenID(t *testing.T) {
	query, err := NewKB("KB-Q", "", "", "", "alpha", "beta")
	if err != nil {
		t.Fatal(err)
	}
	items := []KB{
		query,
		mustGraphKB(t, "KB-Z", "alpha"),
		mustGraphKB(t, "KB-B", "beta", "alpha"),
		mustGraphKB(t, "KB-A", "beta", "alpha"),
		mustGraphKB(t, "KB-X", "outside"),
	}

	got := IDs(Related(items, query))
	want := []string{"KB-A", "KB-B", "KB-Z"}
	if !slices.Equal(got, want) {
		t.Errorf("Related() IDs = %v, want %v", got, want)
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

// TestGraphEdgesCarrySharedTags : chaque arête expose les tags partagés
// (utilisés pour le tooltip du graphe).
func TestGraphEdgesCarrySharedTags(t *testing.T) {
	kbs := sampleGraph()

	edges := GraphEdges(kbs)

	for _, e := range edges {
		if e.A == "KB-0001" && e.B == "KB-0002" {
			if !slices.Equal(e.Tags, []string{"calcul"}) {
				t.Errorf("Tags de l'arête KB-0001↔KB-0002 = %v, want [calcul]", e.Tags)
			}
			return
		}
	}
	t.Fatalf("arête KB-0001↔KB-0002 manquante : %+v", edges)
}

// TestGraphEdgesSharedTagsSorted : plusieurs tags communs → liste triée.
func TestGraphEdgesSharedTagsSorted(t *testing.T) {
	a, err := NewKB("KB-A", "e", "", "", "zeta", "alpha")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	b, err := NewKB("KB-B", "e", "", "", "alpha", "zeta")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	edges := GraphEdges([]KB{a, b})

	if len(edges) != 1 {
		t.Fatalf("GraphEdges() = %d arêtes, want 1", len(edges))
	}
	if !slices.Equal(edges[0].Tags, []string{"alpha", "zeta"}) {
		t.Errorf("Tags partagés = %v, want [alpha zeta]", edges[0].Tags)
	}
}

func TestGraphEdgesPreservesInputPairOrder(t *testing.T) {
	items := []KB{
		mustGraphKB(t, "KB-B", "shared"),
		mustGraphKB(t, "KB-A", "shared"),
		mustGraphKB(t, "KB-C", "shared"),
	}

	got := GraphEdges(items)
	want := []GraphEdge{
		{A: "KB-A", B: "KB-B", Tags: []string{"shared"}},
		{A: "KB-B", B: "KB-C", Tags: []string{"shared"}},
		{A: "KB-A", B: "KB-C", Tags: []string{"shared"}},
	}
	if !slices.EqualFunc(got, want, func(a, b GraphEdge) bool {
		return a.A == b.A && a.B == b.B && slices.Equal(a.Tags, b.Tags)
	}) {
		t.Errorf("GraphEdges() = %+v, want %+v", got, want)
	}
}

func BenchmarkGraphEdges(b *testing.B) {
	items := benchmarkKBs(b, 500)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GraphEdges(items)
	}
}

func BenchmarkGraphEdgesSparse(b *testing.B) {
	items := make([]KB, 1000)
	for i := range items {
		item, err := NewKB(fmt.Sprintf("KB-%04d", i), "", "", "",
			fmt.Sprintf("unique-%04d-a", i), fmt.Sprintf("unique-%04d-b", i))
		if err != nil {
			b.Fatal(err)
		}
		items[i] = item
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GraphEdges(items)
	}
}

func BenchmarkGraphEdgesDense(b *testing.B) {
	items := make([]KB, 500)
	for i := range items {
		items[i] = mustGraphKB(b, fmt.Sprintf("KB-%04d", i), "shared")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GraphEdges(items)
	}
}

func BenchmarkRelated(b *testing.B) {
	items := benchmarkKBs(b, 500)
	query := items[0]
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Related(items, query)
	}
}

func mustGraphKB(t testing.TB, id string, tags ...string) KB {
	t.Helper()
	item, err := NewKB(id, "", "", "", tags...)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func benchmarkKBs(b *testing.B, count int) []KB {
	b.Helper()
	items := make([]KB, count)
	for i := range items {
		item, err := NewKB(fmt.Sprintf("KB-%04d", i), "", "", "",
			fmt.Sprintf("tag-%02d", i%100),
			fmt.Sprintf("tag-%02d", (i+17)%100),
			fmt.Sprintf("tag-%02d", (i+41)%100),
			fmt.Sprintf("tag-%02d", (i+63)%100),
		)
		if err != nil {
			b.Fatal(err)
		}
		items[i] = item
	}
	return items
}

func TestTagStatsEmpty(t *testing.T) {
	if stats := TagStats(nil); len(stats) != 0 {
		t.Errorf("TagStats(nil) = %+v, want vide", stats)
	}
}
