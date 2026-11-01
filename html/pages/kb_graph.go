package pages

import (
	"fmt"
	"grz/internal/kb"
	"strconv"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// graphCSS : styles SVG — survol de nœud discret (le libellé s'éclaircit,
// le point grandit doucement), transitions coupées si l'utilisateur préfère
// les animations réduites.
const graphCSS = `.kg-node circle {
	transform-box: fill-box;
	transform-origin: center;
	transition: transform 150ms ease;
}
.kg-node:hover circle { transform: scale(1.5); }
.kg-node:hover text { fill: #eee8d5; }
@media (prefers-reduced-motion: reduce) {
	.kg-node circle { transition: none; }
}`

// PageKBGraph affiche la vue graphe globale : nœuds = KB, arêtes = tag
// partagé (relation bidirectionnelle), sans librairie externe (SVG inline).
func PageKBGraph(currentPath string, kbs []kb.KB) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("p-6 max-w-3xl mx-auto space-y-4 flex items-center justify-between"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Graphe des KB")),
			A(Href("/"),
				Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-base hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
				gomponents.Text("Retour"),
			),
		),
	}

	if len(kbs) == 0 {
		content = append(content,
			Div(Class("max-w-3xl mx-auto px-6 pb-8"),
				P(Class("text-g-base"), gomponents.Text("Aucune KB : le graphe est vide.")),
			),
		)
		return layout("Graphe des KB", currentPath, content)
	}

	positions := kb.GraphNodes(kbs) // positions en cercle, un seul calcul
	posOf := map[string]kb.GraphNode{}
	for _, n := range positions {
		posOf[n.ID] = n
	}
	byID := map[string]kb.KB{}
	for _, k := range kbs {
		byID[k.ID] = k
	}

	content = append(content,
		Div(Class("max-w-3xl mx-auto px-6 pb-8 space-y-3"),
			Div(Class("bg-g-widget rounded-xl p-8 overflow-x-auto"),
				gomponents.El("svg",
					gomponents.Attr("viewBox", "0 0 1000 1000"),
					gomponents.Attr("class", "w-full h-auto"),
					gomponents.Attr("role", "img"),
					gomponents.Attr("aria-label", "Graphe des KB liées par tags"),

					gomponents.El("style", gomponents.Raw(graphCSS)),
					// Arêtes (derrière les nœuds)
					gomponents.Group(edges(kbs, posOf)),
					// Nœuds cliquables + libellés + tooltips
					gomponents.Group(nodes(positions, byID)),
				),
			),
			// Légende : une phrase, ce que montre le graphe
			P(Class("text-xs text-g-muted"),
				gomponents.Text("Un point par KB, une ligne par tag partagé. Cliquez un point pour ouvrir la fiche."),
			),
		),
	)

	return layout("Graphe des KB", currentPath, content)
}

// edges rend une ligne SVG discrète (filet fin, bouts arrondis) par paire
// de KB partageant un tag, avec le tag partagé en tooltip natif.
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
				gomponents.Attr("stroke", "#28515a"), // filet solarisé (g-sep)
				gomponents.Attr("stroke-width", "1"),
				gomponents.Attr("stroke-opacity", "0.6"),
				gomponents.Attr("stroke-linecap", "round"),
				gomponents.El("title", gomponents.Text(edgeTooltip(e.Tags))),
			),
		)
	}
	return lines
}

// nodes rend chaque KB comme un point discret cliquable (.kg-node), avec un
// libellé mono et un halo sombre (paint-order) pour rester lisible au
// croisement des lignes, plus un tooltip natif (ID, erreur, tags).
func nodes(positions []kb.GraphNode, byID map[string]kb.KB) []gomponents.Node {
	var nodes []gomponents.Node
	for _, n := range positions {
		nodes = append(nodes,
			gomponents.El("a",
				gomponents.Attr("href", "/kb/detail?id="+n.ID),
				gomponents.Attr("class", "kg-node"),
				gomponents.El("title", gomponents.Text(nodeTooltip(byID[n.ID], n.ID))),
				gomponents.El("circle",
					gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
					gomponents.Attr("r", "8"),
					gomponents.Attr("fill", "#2aa198"), // accent cyan solarisé
				),
				gomponents.El("text",
					gomponents.Attr("x", f(n.X)), gomponents.Attr("y", f(n.Y-24)),
					gomponents.Attr("text-anchor", "middle"),
					gomponents.Attr("font-size", "18"),
					gomponents.Attr("font-family", "ui-monospace, monospace"),
					gomponents.Attr("fill", "#839496"), // base0, calme
					// Halo sombre derrière le texte (lisibilité sur les lignes)
					gomponents.Attr("paint-order", "stroke"),
					gomponents.Attr("stroke", "#002b36"),
					gomponents.Attr("stroke-width", "6"),
					gomponents.Attr("stroke-linejoin", "round"),
					gomponents.Text(n.ID),
				),
			),
		)
	}
	return nodes
}

// nodeTooltip construit le tooltip d'un nœud : ID, erreur, tags.
func nodeTooltip(k kb.KB, id string) string {
	msg := firstLine(k.Erreur)
	tags := "aucun tag"
	if len(k.Tags) > 0 {
		tags = "tags : " + strings.Join(k.Tags, ", ")
	}
	if msg == "" {
		return id + " (" + tags + ")"
	}
	return id + " — " + msg + " (" + tags + ")"
}

// edgeTooltip construit le tooltip d'une arête : le(s) tag(s) partagé(s).
func edgeTooltip(tags []string) string {
	if len(tags) == 1 {
		return fmt.Sprintf("Tag partagé : %s", tags[0])
	}
	return fmt.Sprintf("Tags partagés : %s", strings.Join(tags, ", "))
}

// f formate une coordonnée flottante pour un attribut SVG.
func f(v float64) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}
