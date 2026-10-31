package kb

import (
	"math"
	"slices"
	"strings"
)

// TagStat est l'inventaire d'un tag et le nombre de KB qui le portent.
type TagStat struct {
	Name  string
	Count int
}

// IDs retourne les identifiants d'une liste de KB, triés.
func IDs(kbs []KB) []string {
	ids := make([]string, 0, len(kbs))
	for _, k := range kbs {
		ids = append(ids, k.ID)
	}
	slices.Sort(ids)
	return ids
}

// TagStats calcule les tags avec leur nombre de KB, triés par fréquence
// décroissante puis par nom.
func TagStats(kbs []KB) []TagStat {
	counts := map[string]int{}
	for _, k := range kbs {
		for _, tag := range k.Tags {
			counts[tag]++
		}
	}

	stats := []TagStat{}
	for name, count := range counts {
		stats = append(stats, TagStat{Name: name, Count: count})
	}
	slices.SortFunc(stats, func(a, b TagStat) int {
		if a.Count != b.Count {
			return b.Count - a.Count
		}
		return strings.Compare(a.Name, b.Name)
	})

	return stats
}

// FilterByTag retourne les KB portant le tag tag, triées par ID.
func FilterByTag(kbs []KB, tag string) []KB {
	filtered := []KB{}
	for _, k := range kbs {
		if slices.Contains(k.Tags, tag) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

// Related retourne les KB partageant au moins un tag avec kb (relation
// symétrique : A↔B), triées par nombre de tags communs puis ID.
func Related(kbs []KB, kb KB) []KB {
	type candidate struct {
		k      KB
		common int
	}

	var candidates []candidate
	for _, other := range kbs {
		if other.ID == kb.ID {
			continue
		}
		common := 0
		for _, tag := range kb.Tags {
			if slices.Contains(other.Tags, tag) {
				common++
			}
		}
		if common > 0 {
			candidates = append(candidates, candidate{other, common})
		}
	}

	slices.SortFunc(candidates, func(a, b candidate) int {
		if a.common != b.common {
			return b.common - a.common
		}
		return strings.Compare(a.k.ID, b.k.ID)
	})

	related := make([]KB, 0, len(candidates))
	for _, c := range candidates {
		related = append(related, c.k)
	}
	return related
}

// Graph constants pour la disposition en cercle (viewBox 1000x1000).
const (
	graphCenterX = 500.0
	graphCenterY = 500.0
	graphRadius  = 380.0
)

// GraphNode est un nœud du graphe, positionné sur un cercle.
type GraphNode struct {
	ID string
	X  float64
	Y  float64
}

// GraphEdge est une arête entre deux KB partageant au moins un tag
// (relation bidirectionnelle : l'ordre A/B est normalisé par ID croissant).
type GraphEdge struct {
	A string
	B string
}

// GraphNodes place chaque KB sur un cercle, triées par ID.
func GraphNodes(kbs []KB) []GraphNode {
	sorted := slices.Clone(kbs)
	slices.SortFunc(sorted, func(a, b KB) int { return strings.Compare(a.ID, b.ID) })

	nodes := make([]GraphNode, 0, len(sorted))
	for i, k := range sorted {
		angle := 2 * math.Pi * float64(i) / float64(len(sorted))
		nodes = append(nodes, GraphNode{
			ID: k.ID,
			X:  graphCenterX + graphRadius*math.Cos(angle),
			Y:  graphCenterY + graphRadius*math.Sin(angle),
		})
	}
	return nodes
}

// GraphEdges retourne les arêtes entre KB partageant au moins un tag.
func GraphEdges(kbs []KB) []GraphEdge {
	var edges []GraphEdge

	for i, a := range kbs {
		for _, b := range kbs[i+1:] {
			if sharesTag(a, b) {
				e := GraphEdge{A: a.ID, B: b.ID}
				if e.A > e.B {
					e.A, e.B = e.B, e.A
				}
				edges = append(edges, e)
			}
		}
	}
	return edges
}

func sharesTag(a, b KB) bool {
	for _, tag := range a.Tags {
		if slices.Contains(b.Tags, tag) {
			return true
		}
	}
	return false
}

// ByID retourne un pointeur sur la KB id dans la liste, ou nil si absente.
func ByID(kbs []KB, id string) *KB {
	for i := range kbs {
		if kbs[i].ID == id {
			return &kbs[i]
		}
	}
	return nil
}
