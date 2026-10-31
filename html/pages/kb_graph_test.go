package pages

import (
	"strings"
	"testing"

	"grz/internal/kb"
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
	if !strings.Contains(html, `paint-order="stroke"`) {
		t.Errorf("les libellés doivent utiliser paint-order pour le halo")
	}
	// Nœuds stylables / survolables
	if !strings.Contains(html, "kg-node") {
		t.Errorf("classe kg-node absente : pas de survol possible")
	}
	// Arêtes discrètes
	if !strings.Contains(html, `stroke-width="1"`) {
		t.Errorf("les arêtes doivent faire stroke-width=1")
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
