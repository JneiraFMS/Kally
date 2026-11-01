package pages

import (
	"fmt"
	kb "grz/internal/domain/kb"
	"math"
	"slices"
	"strconv"
	"strings"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// graphCSS : styles SVG — survol de nœud discret (le libellé s'éclaircit,
// le point grandit doucement), arêtes liées accentuées en jaune (kg-hot)
// quand le pointeur survole un nœud, transitions coupées si l'utilisateur
// préfère les animations réduites.
// graphCSS : styles SVG — survol de nœud discret (le libellé s'éclaircit,
// le point grandit doucement), arêtes liées accentuées en jaune (kg-hot)
// quand le pointeur survole un nœud, transitions coupées si l'utilisateur
// préfère les animations réduites. Couleurs via var(--g-*) : le même SVG
// se colore selon le thème actif (dark par défaut / [data-theme=light]).
// Les attributs de présentation SVG ne supportant pas var(), toutes les
// couleurs passent par ces classes CSS.
const graphCSS = `.kg-edge {
	stroke: var(--g-sep);
	stroke-opacity: 0.6;
	stroke-width: 1;
	stroke-linecap: round;
}
.kg-dot { fill: var(--g-cyan); }
.kg-label {
	fill: var(--g-base);
	paint-order: stroke;
	stroke: var(--g-bg);
	stroke-width: 6;
	stroke-linejoin: round;
}
.kg-node circle {
	transform-box: fill-box;
	transform-origin: center;
	transition: transform 150ms ease;
}
.kg-node { cursor: pointer; }
.kg-node:hover circle { transform: scale(1.5); }
.kg-node:hover .kg-label { fill: var(--g-hi); }
.kg-hot {
	stroke: var(--g-accent) !important;
	stroke-opacity: 1 !important;
	stroke-width: 2.5 !important;
}
@media (prefers-reduced-motion: reduce) {
	.kg-node circle { transition: none; }
}`

const (
	graphCenterX = 500.0
	graphCenterY = 500.0
	graphRadius  = 380.0
)

type graphNode struct {
	ID string
	X  float64
	Y  float64
}

// graphNodes calculates SVG-specific circular coordinates for the graph view.
func graphNodes(kbs []kb.KB) []graphNode {
	sorted := slices.Clone(kbs)
	slices.SortFunc(sorted, func(a, b kb.KB) int { return strings.Compare(a.ID, b.ID) })
	nodes := make([]graphNode, 0, len(sorted))
	for i, item := range sorted {
		angle := 2 * math.Pi * float64(i) / float64(len(sorted))
		nodes = append(nodes, graphNode{
			ID: item.ID,
			X:  graphCenterX + graphRadius*math.Cos(angle),
			Y:  graphCenterY + graphRadius*math.Sin(angle),
		})
	}
	return nodes
}

// graphScript : surlignage des arêtes du nœud survolé. Chaque nœud porte
// data-id, chaque arête data-a/data-b : au survol (délégation par
// événements souris), les arêtes liées reçoivent la classe kg-hot sans
// état global ni re-rendu.
const graphScript = `/* kg-highlight-js : surlignage arêtes du nœud survolé */
(function () {
"use strict";
var lines = document.querySelectorAll("svg line[data-a]");
function setHot(id, on) {
	for (var i = 0; i < lines.length; i++) {
		var l = lines[i];
		if (l.getAttribute("data-a") === id || l.getAttribute("data-b") === id) {
			l.classList.toggle("kg-hot", on);
		}
	}
}
var nodes = document.querySelectorAll(".kg-node");
for (var i = 0; i < nodes.length; i++) {
	(function (node) {
		node.addEventListener("mouseenter", function () { setHot(node.getAttribute("data-id"), true); });
		node.addEventListener("mouseleave", function () { setHot(node.getAttribute("data-id"), false); });
	})(nodes[i]);
}
})();`

// PageKBGraph affiche la vue graphe globale : nœuds = KB, arêtes = tag
// partagé (relation bidirectionnelle), sans librairie externe (SVG inline).
func PageKBGraph(currentPath string, kbs []kb.KB) gomponents.Node {
	content := []gomponents.Node{
		Div(Class("max-w-5xl mx-auto px-6 py-6 space-y-4 flex items-center justify-between"),
			H1(Class("text-lg font-semibold text-g-hi"), gomponents.Text("Graphe des KB")),
			A(Href("/"),
				Class("px-3.5 py-1.5 rounded-lg bg-g-block text-g-base hover:text-g-hi hover:bg-g-hover text-xs transition-colors"),
				gomponents.Text("Retour"),
			),
		),
	}

	if len(kbs) == 0 {
		content = append(content,
			Div(Class("max-w-5xl mx-auto px-6 pb-8"),
				P(Class("text-g-base"), gomponents.Text("Aucune KB : le graphe est vide.")),
			),
		)
		return layout("Graphe des KB", currentPath, content)
	}

	positions := graphNodes(kbs) // positions en cercle, un seul calcul
	posOf := map[string]graphNode{}
	for _, n := range positions {
		posOf[n.ID] = n
	}
	byID := map[string]kb.KB{}
	for _, k := range kbs {
		byID[k.ID] = k
	}

	content = append(content,
		Div(Class("max-w-5xl mx-auto px-6 pb-8 space-y-3"),
			Div(Class("max-w-3xl bg-g-widget rounded-xl p-8 overflow-x-auto"),
				gomponents.El("svg",
					gomponents.Attr("viewBox", "0 0 1000 1000"),
					gomponents.Attr("class", "w-full h-auto"),
					gomponents.Attr("role", "img"),
					gomponents.Attr("aria-label", "Graphe des KB liées par tags"),

					gomponents.El("style", gomponents.Raw(graphCSS)),
					// Arêtes (derrière les nœuds), liées à leurs KB via
					// data-a/data-b pour le surlignage au survol.
					gomponents.Group(edges(kbs, posOf)),
					// Nœuds cliquables + libellés + tooltips
					gomponents.Group(nodes(positions, byID)),
				),
			),
			// Légende : une phrase, ce que montre le graphe
			P(Class("text-xs text-g-base"),
				gomponents.Text("Un point par KB, une ligne par tag partagé. Survolez un point pour suivre ses liens, cliquez-le pour ouvrir la fiche."),
			),
			// Surlignage des arêtes du nœud survolé (dégradation douce sans JS)
			gomponents.El("script", gomponents.Raw(graphScript)),
		),
	)

	return layout("Graphe des KB", currentPath, content)
}

// edges rend une ligne SVG discrète (filet fin, bouts arrondis) par paire
// de KB partageant un tag, avec le tag partagé en tooltip natif.
func edges(kbs []kb.KB, posOf map[string]graphNode) []gomponents.Node {
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
				gomponents.Attr("class", "kg-edge"),
				// Extrémités pour le surlignage kg-hot au survol du nœud
				gomponents.Attr("data-a", e.A),
				gomponents.Attr("data-b", e.B),
				gomponents.El("title", gomponents.Text(edgeTooltip(e.Tags))),
			),
		)
	}
	return lines
}

// nodes rend chaque KB comme un point discret cliquable (.kg-node) : une
// zone de clic élargie (cercle transparent r=20, cible ≥ 24 px une fois le
// viewBox ramené à ~700 px, WCAG 2.5.8), un point coloré, un libellé mono
// (posé SOUS le point pour la moitié basse du cercle, au-dessus sinon, pour
// éviter les chevauchements), un halo sombre (paint-order) et un tooltip
// natif (ID, erreur, tags). data-id relie le nœud au surlignage des arêtes.
func nodes(positions []graphNode, byID map[string]kb.KB) []gomponents.Node {
	var nodes []gomponents.Node
	for _, n := range positions {
		labelY := n.Y - 24
		if n.Y > 500 { // moitié basse : libellé sous le point
			labelY = n.Y + 40
		}
		nodes = append(nodes,
			gomponents.El("a",
				gomponents.Attr("href", "/kb/detail?id="+n.ID),
				gomponents.Attr("class", "kg-node"),
				gomponents.Attr("data-id", n.ID),
				gomponents.El("title", gomponents.Text(nodeTooltip(byID[n.ID], n.ID))),
				// Zone de clic élargie (transparente, tactile inclus)
				gomponents.El("circle",
					gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
					gomponents.Attr("r", "20"),
					gomponents.Attr("fill", "transparent"),
				),
				gomponents.El("circle",
					gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
					gomponents.Attr("r", "8"),
					gomponents.Attr("class", "kg-dot"),
				),
				gomponents.El("text",
					gomponents.Attr("x", f(n.X)), gomponents.Attr("y", f(labelY)),
					gomponents.Attr("text-anchor", "middle"),
					gomponents.Attr("font-size", "18"),
					gomponents.Attr("font-family", "ui-monospace, monospace"),
					gomponents.Attr("class", "kg-label"),
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
