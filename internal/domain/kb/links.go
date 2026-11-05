package kb

import (
	"slices"
	"strings"

	"grz/internal/domain/wikilink"
)

// LinkTargets retourne les KB citées explicitement avec [[...]] dans les
// sections markdown d'une KB (Notes puis Résolution), dans l'ordre
// d'apparition et sans doublon. La section Erreur est une trace brute, non
// liée ; le code Markdown est ignoré par le parseur wikilink.
func LinkTargets(k KB) []string {
	content := k.Notes + "\n" + k.Resolution
	targets := []string{}
	for _, target := range wikilink.Parse(content) {
		id := string(target)
		if err := ValidateID(id); err != nil {
			continue
		}
		targets = append(targets, id)
	}
	return targets
}

// Backlinks retourne les KB qui citent id avec [[...]], hors la KB
// elle-même, triées par ID. La relation est orientée : A cite B ne signifie
// pas que B cite A ; B affiche donc A en rétrolien sans réciprocité.
func Backlinks(kbs []KB, id string) []KB {
	backlinks := []KB{}
	for _, k := range kbs {
		if k.ID == id {
			continue
		}
		if slices.Contains(LinkTargets(k), id) {
			backlinks = append(backlinks, k)
		}
	}
	slices.SortFunc(backlinks, func(a, b KB) int {
		return strings.Compare(a.ID, b.ID)
	})
	return backlinks
}

// Neighbors retourne les KB voisines d'une KB pour le mini-graphe : celles
// qui partagent au moins un tag et celles reliées explicitement par [[...]],
// dans un sens ou dans l'autre. La KB elle-même est exclue ; le résultat est
// trié par ID. Une KB inconnue n'a aucun voisin.
func Neighbors(kbs []KB, id string) []KB {
	self := ByID(kbs, id)
	if self == nil {
		return []KB{}
	}

	selfTags := make(map[string]struct{}, len(self.Tags))
	for _, tag := range self.Tags {
		selfTags[tag] = struct{}{}
	}
	selfLinks := make(map[string]struct{})
	for _, target := range LinkTargets(*self) {
		selfLinks[target] = struct{}{}
	}

	neighbors := []KB{}
	for _, k := range kbs {
		if k.ID == id {
			continue
		}
		sharedTag := false
		for _, tag := range k.Tags {
			if _, ok := selfTags[tag]; ok {
				sharedTag = true
				break
			}
		}
		linked := slices.Contains(LinkTargets(k), id)
		if _, ok := selfLinks[k.ID]; ok {
			linked = true
		}
		if sharedTag || linked {
			neighbors = append(neighbors, k)
		}
	}

	slices.SortFunc(neighbors, func(a, b KB) int {
		return strings.Compare(a.ID, b.ID)
	})
	return neighbors
}

// LinkEdges retourne les arêtes explicites entre KB existantes : une arête
// non orientée (A<B) par paire, dédupliquée quel que soit le sens de la
// citation. Les cibles absentes et les auto-références sont ignorées.
// Le résultat est trié par A puis B (affichage déterministe du graphe).
func LinkEdges(kbs []KB) []GraphEdge {
	known := make(map[string]struct{}, len(kbs))
	for _, k := range kbs {
		known[k.ID] = struct{}{}
	}

	type pair struct{ a, b string }
	seen := make(map[pair]struct{})
	edges := []GraphEdge{}
	for _, k := range kbs {
		for _, target := range LinkTargets(k) {
			if target == k.ID {
				continue
			}
			if _, ok := known[target]; !ok {
				continue
			}
			a, b := k.ID, target
			if a > b {
				a, b = b, a
			}
			key := pair{a: a, b: b}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, GraphEdge{A: a, B: b, Linked: true})
		}
	}

	slices.SortFunc(edges, func(a, b GraphEdge) int {
		if a.A != b.A {
			return strings.Compare(a.A, b.A)
		}
		return strings.Compare(a.B, b.B)
	})
	return edges
}
