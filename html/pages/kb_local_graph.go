package pages

import (
	"math"
	"slices"
	"strings"

	kb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
)

// Graphique voisin limité : au-delà, le graphe global reste l'outil de
// navigation (label « …et N autres » sous le mini-graphe).
const localGraphLimit = 8

// localGraphCSS reprend le style du graphe global (kg-*) et ajoute la
// classe du nœud central : accent, comme la note actuellement ouverte dans
// Logseq. Couleurs via var(--g-*) : suit le thème actif.
const localGraphCSS = graphCSS + `
.kg-main { fill: var(--g-accent); }`

// localGraph rend le mini-graphe d'une note (inspiré Logseq) : la KB
// au centre (accent), ses voisins placés en cercle, arêtes = tags partagés
// à l'intérieur du sous-ensemble. Chaque voisin est cliquable vers sa fiche.
// Sans voisin, la fonction ne retourne rien (la section « KB liées »
// annonce déjà le vide).
func localGraph(principal kb.KB, related []kb.KB) gomponents.Node {
	if len(related) == 0 {
		return nil
	}
	voisins := related
	if len(voisins) > localGraphLimit {
		voisins = voisins[:localGraphLimit]
	}

	// Positions déterministes : voisin trié par ID autour du centre.
	order := slices.Clone(voisins)
	slices.SortFunc(order, func(a, b kb.KB) int { return strings.Compare(a.ID, b.ID) })

	const (
		cote   = 440.0
		centre = cote / 2
		rayon  = 150.0
	)

	posOf := map[string]graphNode{principal.ID: {ID: principal.ID, X: centre, Y: centre}}
	for i, n := range order {
		angle := 2*math.Pi*float64(i)/float64(len(order)) - math.Pi/2
		posOf[n.ID] = graphNode{
			ID: n.ID,
			X:  centre + rayon*math.Cos(angle),
			Y:  centre + rayon*math.Sin(angle),
		}
	}

	// Arêtes internes au sous-ensemble (principal ↔ voisins, voisins ↔
	// voisins) : même relation que le graphe global, portée locale.
	var lines []gomponents.Node
	for _, e := range kb.GraphEdges(append([]kb.KB{principal}, voisins...)) {
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
				gomponents.El("title", gomponents.Text(edgeTooltip(e.Tags))),
			),
		)
	}

	nodes := []gomponents.Node{localNode(principal.ID, posOf[principal.ID], true)}
	for _, n := range order {
		nodes = append(nodes, localNode(n.ID, posOf[n.ID], false))
	}

	surplus := len(related) - localGraphLimit

	return gomponents.El("svg",
		gomponents.Attr("viewBox", "0 0 440 440"),
		gomponents.Attr("class", "w-full h-auto"),
		gomponents.Attr("role", "img"),
		gomponents.Attr("aria-label", "Mini-graphe : "+principal.ID+" et ses KB liées"),
		gomponents.Attr("data-kg", "local"),
		gomponents.El("style", gomponents.Raw(localGraphCSS)),
		gomponents.Group(lines),
		gomponents.Group(nodes),
		gomponents.If(surplus > 0,
			gomponents.El("text",
				gomponents.Attr("x", f(centre)), gomponents.Attr("y", f(cote-8)),
				gomponents.Attr("text-anchor", "middle"),
				gomponents.Attr("font-size", "14"),
				gomponents.Attr("class", "kg-label"),
				gomponents.Text("… et plus dans le graphe global"),
			),
		),
	)
}

// localNode rend un nœud du mini-graphe : accessible (zone de clic ≥ 24 px),
// libellé mono, tooltip natif. Le nœud principal (isMain=true) porte la
// classe kg-main (accent) ; les voisins sont cliquables vers leur fiche.
func localNode(id string, n graphNode, isMain bool) gomponents.Node {
	dotClass := "kg-dot"
	if isMain {
		dotClass = "kg-dot kg-main"
	}
	href := "/kb/detail?id=" + id
	if isMain {
		// La note courante est visible : inutile de relier vers elle-même
		href = serviceNowURL(id)
	}
	return gomponents.El("a",
		gomponents.Attr("href", href),
		gomponents.Attr("class", "kg-node"),
		gomponents.Attr("data-id", id),
		gomponents.El("title", gomponents.Text(id)),
		gomponents.El("circle",
			gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
			gomponents.Attr("r", "20"),
			gomponents.Attr("fill", "transparent"),
		),
		gomponents.El("circle",
			gomponents.Attr("cx", f(n.X)), gomponents.Attr("cy", f(n.Y)),
			gomponents.Attr("r", "10"),
			gomponents.Attr("class", dotClass),
		),
		gomponents.El("text",
			gomponents.Attr("x", f(n.X)), gomponents.Attr("y", f(n.Y-34)),
			gomponents.Attr("text-anchor", "middle"),
			gomponents.Attr("font-size", "16"),
			gomponents.Attr("font-family", "ui-monospace, monospace"),
			gomponents.Attr("class", "kg-label"),
			gomponents.Text(id),
		),
	)
}
