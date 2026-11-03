package pages

import (
	"strings"
	"testing"

	kb "grz/internal/domain/kb"
)

// TestPageKBGraphStyle : style minimaliste du graphe — arêtes fines,
// libellés avec halo sombre (paint-order), nœuds en .kg-node survolables,
// tooltips natifs <title> et légende explicative.
func TestPageKBGraphStyle(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}
	html := renderedPage(t, PageKBGraph("/kb/graph", kbs))

	// Tooltips natifs (nœud : ID + erreur + tags ; arête : tag partagé)
	if !strings.Contains(html, "<title>") {
		t.Errorf("les nœuds/arêtes doivent avoir un tooltip <title>")
	}
	if !strings.Contains(html, "cc_flow") {
		t.Errorf("le tooltip de l'arête doit mentionner le tag partagé")
	}
	// Halo sombre derrière les libellés (lisibilité au croisement des lignes)
	if !strings.Contains(html, "paint-order: stroke") {
		t.Errorf("les libellés doivent utiliser paint-order pour le halo")
	}
	// Nœuds stylables / survolables
	if !strings.Contains(html, "kg-node") {
		t.Errorf("classe kg-node absente : pas de survol possible")
	}
	// Arêtes discrètes
	if !strings.Contains(html, "stroke-width: 1") {
		t.Errorf("les arêtes doivent faire stroke-width=1 (classe kg-edge)")
	}
	// Arêtes/nœuds/libellés couleur par classe CSS (thèmes via var())
	if !strings.Contains(html, ".kg-edge") || !strings.Contains(html, ".kg-dot") || !strings.Contains(html, ".kg-label") {
		t.Errorf("classes de couleur du graphe manquantes (kg-edge/kg-dot/kg-label)")
	}
	if strings.Contains(html, `stroke="#`) || strings.Contains(html, `fill="#`) {
		t.Errorf("couleurs SVG en attribut hex : impossible à thématiser")
	}
	// Légende sous le graphe
	if !strings.Contains(html, "Un point par KB") {
		t.Errorf("légende du graphe manquante")
	}
	// Motion réduit respecté
	if !strings.Contains(html, "prefers-reduced-motion") {
		t.Errorf("la feuille SVG doit respecter prefers-reduced-motion")
	}
}

// TestPageKBGraphTooltipNode : le tooltip nœud contient ID, erreur et tags.
func TestPageKBGraphTooltipNode(t *testing.T) {
	k := mustCardKB(t, "KB06644754", "cc_flow")
	html := renderedPage(t, PageKBGraph("/kb/graph", []kb.KB{k}))

	if !strings.Contains(html, "KB06644754 — erreur démo") {
		t.Errorf("tooltip nœud sans ID/erreur : attendu %q", "KB06644754 — erreur démo")
	}
	if !strings.Contains(html, "tags : cc_flow") {
		t.Errorf("tooltip nœud sans tags")
	}
}

// TestPageKBGraphHitTargets : chaque nœud a une zone de clic élargie
// (cercle de hit transparent, r=20 en un viewBox affiché ~700px ≈ 28 px réels,
// ≥ la cible minimale WCAG 2.5.8 de 24 px), et un curseur dédié.
func TestPageKBGraphHitTargets(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow"),
	}
	html := renderedPage(t, PageKBGraph("/kb/graph", kbs))

	if strings.Count(html, `<circle cx=`) < 2 || !strings.Contains(html, `r="20"`) {
		t.Errorf("zones de clic élargies (cercle r=20) manquantes:\n%s", html)
	}
	if !strings.Contains(html, "cursor:") {
		t.Errorf("le graphe doit être cliquable (cursor sur kg-node)")
	}
}

// TestPageKBGraphEdgeHighlight : les arêtes sont associées à leurs deux KB
// (data-a/data-b) et un script les met en surbrillance au survol d'un nœud
// (classe kg-hot accentuée dans la feuille SVG).
func TestPageKBGraphEdgeHighlight(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow"),
	}
	html := renderedPage(t, PageKBGraph("/kb/graph", kbs))

	if !strings.Contains(html, `data-a="KB06644754"`) || !strings.Contains(html, `data-b="KB22547"`) {
		t.Errorf("arêtes sans data-a/data-b : le surlignage est impossible:\n%s", html)
	}
	if !strings.Contains(html, "kg-hot") {
		t.Errorf("règle CSS kg-hot (surlignage arêtes) manquante")
	}
	if !strings.Contains(html, `data-id="KB06644754"`) {
		t.Errorf("nœuds sans data-id : le surlignage ne peut pas cibler le nœud survolé")
	}
	if !strings.Contains(html, "kg-highlight-js") {
		t.Errorf("script de surlignage manquant")
	}
}

// TestPageKBGraphLabelSide : les libellés des nœuds de la moitié basse du
// cercle passent SOUS le point (y+40) pour éviter le chevauchement dans la
// zone densément reliée ; ceux du haut restent au-dessus (y−24).
func TestPageKBGraphLabelSide(t *testing.T) {
	// 3 KB → angles 0, 2π/3, 4π/3 (R=380, centre 500) :
	//   KB-1 : (880.0, 500.0) → label à y = 476.0 (au-dessus)
	//   KB-2 : (310.0, 829.1) → label à y = 869.1 (en-dessous, moitié basse)
	//   KB-3 : (310.0, 170.9) → label à y = 146.9 (au-dessus)
	kbs := []kb.KB{
		mustCardKB(t, "KB-1"),
		mustCardKB(t, "KB-2"),
		mustCardKB(t, "KB-3"),
	}
	html := renderedPage(t, PageKBGraph("/kb/graph", kbs))

	if !strings.Contains(html, `y="869.1"`) {
		t.Errorf("KB-2 (moitié basse) : label attendu sous le nœud (y=869.1):\n%s", html)
	}
	if !strings.Contains(html, `y="476.0"`) {
		t.Errorf("KB-1 (moitié haute) : label attendu au-dessus du nœud (y=476.0)")
	}
	if !strings.Contains(html, `y="146.9"`) {
		t.Errorf("KB-3 (moitié haute) : label attendu au-dessus du nœud (y=146.9)")
	}
}

// TestPageKBGraphHeaderAlignment : l'en-tête est une rangée flex dont les
// deux enfants sont centrés verticalement. `space-y-4` y ajoutait une marge
// haute au lien « Retour » (8 px de décalage avec le titre) : la rangée ne
// doit pas mélanger l'espacement vertical empilé et l'alignement centré.
func TestPageKBGraphHeaderAlignment(t *testing.T) {
	html := renderedPage(t, PageKBGraph("/kb/graph", nil))

	if strings.Contains(html, "space-y-4 flex") {
		t.Errorf("en-tête du graphe : space-y-4 sur une rangée flex (décalage vertical du lien Retour)")
	}
	if !strings.Contains(html, "flex items-center justify-between") {
		t.Errorf("en-tête du graphe sans alignement centré:\n%s", html)
	}
}
