package kb

import (
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

	queryTags := make(map[string]struct{}, len(kb.Tags))
	for _, tag := range kb.Tags {
		queryTags[tag] = struct{}{}
	}

	candidates := make([]candidate, 0)
	for _, other := range kbs {
		if other.ID == kb.ID {
			continue
		}
		common := 0
		for _, tag := range other.Tags {
			if _, ok := queryTags[tag]; ok {
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

// GraphEdge est une arête entre deux KB partageant au moins un tag
// (relation bidirectionnelle : l'ordre A/B est normalisé par ID croissant).
// Tags contient les tags communs aux deux KB, triés (affichés en tooltip).
type GraphEdge struct {
	A    string
	B    string
	Tags []string
}

// GraphEdges retourne les arêtes entre KB partageant au moins un tag,
// avec les tags partagés. Les arêtes suivent l'ordre des paires d'entrée.
func GraphEdges(kbs []KB) []GraphEdge {
	type pair struct{ first, second int }

	indicesByTag := make(map[string][]int)
	var candidatePairs uint64
	for i, item := range kbs {
		for _, tag := range item.Tags {
			posting := indicesByTag[tag]
			candidatePairs += uint64(len(posting))
			indicesByTag[tag] = append(posting, i)
		}
	}
	possiblePairs := uint64(len(kbs)) * uint64(max(len(kbs)-1, 0)) / 2
	if candidatePairs >= possiblePairs/2 {
		return graphEdgesDense(kbs)
	}

	// Sparse graphs are built from tag postings, avoiding an O(n²) scan.
	sharedByPair := make(map[pair][]string)
	for tag, indices := range indicesByTag {
		for i, first := range indices {
			for _, second := range indices[i+1:] {
				key := pair{first: first, second: second}
				sharedByPair[key] = append(sharedByPair[key], tag)
			}
		}
	}
	pairs := make([]pair, 0, len(sharedByPair))
	for key := range sharedByPair {
		pairs = append(pairs, key)
	}
	slices.SortFunc(pairs, func(a, b pair) int {
		if a.first != b.first {
			return a.first - b.first
		}
		return a.second - b.second
	})

	edges := make([]GraphEdge, 0, len(pairs))
	for _, key := range pairs {
		common := sharedByPair[key]
		slices.Sort(common)
		edge := GraphEdge{A: kbs[key.first].ID, B: kbs[key.second].ID, Tags: common}
		if edge.A > edge.B {
			edge.A, edge.B = edge.B, edge.A
		}
		edges = append(edges, edge)
	}
	return edges
}

func graphEdgesDense(kbs []KB) []GraphEdge {
	tagSets := make([]map[string]struct{}, len(kbs))
	for i, item := range kbs {
		if len(item.Tags) == 0 {
			continue
		}
		tagSets[i] = make(map[string]struct{}, len(item.Tags))
		for _, tag := range item.Tags {
			tagSets[i][tag] = struct{}{}
		}
	}

	edges := make([]GraphEdge, 0)
	for i, a := range kbs {
		for j := i + 1; j < len(kbs); j++ {
			var common []string
			for _, tag := range a.Tags {
				if _, ok := tagSets[j][tag]; ok {
					common = append(common, tag)
				}
			}
			if len(common) == 0 {
				continue
			}
			slices.Sort(common)
			edge := GraphEdge{A: a.ID, B: kbs[j].ID, Tags: common}
			if edge.A > edge.B {
				edge.A, edge.B = edge.B, edge.A
			}
			edges = append(edges, edge)
		}
	}
	return edges
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
