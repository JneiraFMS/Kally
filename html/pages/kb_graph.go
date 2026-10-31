package pages

import (
	"grz/internal/kb"
	"strconv"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// PageKBGraph affiche la vue graphe globale : nœuds = KB, arêtes = tag
// partagé (relation bidirectionnelle), sans librairie externe (SVG inline).
func PageKBGraph(currentPath string, kbs []kb.KB) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("p-8 max-w-3xl mx-auto space-y-4 flex items-center justify-between"),
			H1(Class("text-2xl font-bold"), gomponents.Text("Graphe des KB")),
			A(Href("/kb"),
				Class("px-3 py-1.5 rounded-md border border-gray-200 bg-white text-sm text-gray-600 hover:bg-gray-100"),
				gomponents.Text("← Liste"),
			),
		),
	}

	if len(kbs) == 0 {
		content = append(content,
			Div(Class("max-w-3xl mx-auto px-8 pb-8"),
				P(Class("text-gray-500"), gomponents.Text("Aucune KB : le graphe est vide.")),
			),
		)
		return layout("Graphe des KB", currentPath, content)
	}

	positions := kb.GraphNodes(kbs) // positions en cercle, une seul calcul
	posOf := map[string]kb.GraphNode{}
	for _, n := range positions {
		posOf[n.ID] = n
	}

	content = append(content,
		Div(Class("max-w-3xl mx-auto px-8 pb-8"),
			Div(Class("bg-white ring-1 ring-gray-200 rounded-lg shadow-sm p-6 overflow-x-auto"),
				gomponents.El("svg",
					gomponents.Attr("viewBox", "0 0 1000 1000"),
					gomponents.Attr("class", "w-full h-auto"),
					gomponents.Attr("role", "img"),
					gomponents.Attr("aria-label", "Graphe des KB liées par tags"),

					// Arêtes (derrière les nœuds)
					gomponents.Group(edges(kbs, posOf)),
					// Nœuds cliquables + libellés
					gomponents.Group(nodes(positions)),
				),
			),
		),
	)

	return layout("Graphe des KB", currentPath, content)
}

// edges rend une ligne SVG par paire de KB partageant un tag.
func edges(kbs []kb.KB, posOf map[string]kb.GraphNode) []gomponents.Node {
	var lines []gomponents.Node
	for _, e := range kb.GraphEdges(kbs) {
		a, okA := posOf[e.A]
		b, okB := posOf[e.B]
		if !okA || !okB {
			continue
		}
		lines = append(lines,
			gomponents.El("line",
				gomponents.Attr("x1", f(a.X)), gomponents.Attr("y1", f(a.Y)),
				gomponents.Attr("x2", f(b.X)), gomponents.Attr("y2", f(b.Y)),
				gomponents.Attr("stroke", "#d1d5db"),
				gomponents.Attr("stroke-width", "2"),
			),
		)
	}
	return lines
}

// nodes rend chaque KB comme un point cliquable avec son libellé.
func nodes(positions []kb.GraphNode) []gomponents.Node {
	var nodes []gomponents.Node
	for _, n := range positions {
		nodes = append(nodes,
			gomponents.El("a",
				gomponents.Attr("href", "/kb/detail?id="+n.ID),
				gomponents.El("circle",
					gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
					gomponents.Attr("r", "14"),
					gomponents.Attr("fill", "#3b82f6"),
				),
				gomponents.El("text",
					gomponents.Attr("x", f(n.X)), gomponents.Attr("y", f(n.Y-24)),
					gomponents.Attr("text-anchor", "middle"),
					gomponents.Attr("font-size", "22"),
					gomponents.Attr("fill", "#374151"),
					gomponents.Text(n.ID),
				),
			),
		)
	}
	return nodes
}

// f formate une coordonnée flottante pour un attribut SVG.
func f(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}
