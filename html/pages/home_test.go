package pages

import (
	"strings"
	"testing"

	"grz/internal/kb"

	"maragu.dev/gomponents"
)

// noNestedAnchors vérifie qu'aucun lien <a> n'est imbriqué dans un autre
// (HTML invalide : le navigateur referme le premier <a> et casse le DOM).
func noNestedAnchors(t *testing.T, html string) {
	t.Helper()

	depth := 0
	for i := 0; i < len(html); i++ {
		switch {
		case hasAnchorOpen(html[i:]):
			depth++
			if depth > 1 {
				start := max(0, i-60)
				end := min(len(html), i+60)
				t.Fatalf("un <a> est imbriqué dans un autre <a> (profondeur %d) autour de : ...%s...", depth, html[start:end])
			}
			i += 2 // "<a" ; l'appelant continue au caractère suivant
		case strings.HasPrefix(html[i:], "</a>"):
			if depth > 0 {
				depth--
			}
			i += 3
		}
	}
}

// hasAnchorOpen détecte "<a " ou "<a>" suivi d'un possible contenu.
func hasAnchorOpen(s string) bool {
	if strings.HasPrefix(s, "<a ") || strings.HasPrefix(s, "<a>") {
		return true
	}
	return false
}

func renderedPage(t *testing.T, node gomponents.Node) string {
	t.Helper()
	return renderNode(t, node)
}

func mustCardKB(t *testing.T, id string, tags ...string) kb.KB {
	t.Helper()
	k, err := kb.NewKB(id, "erreur démo", "notes", "résolution", tags...)
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	return k
}

// TestLayoutInjectsGlancePalette : le thème doit être chargé
// (config Tailwind inline contenant la palette g-*), à la Solarized :
// fond #002b36 (base03) + accent jaune délavé #b58900.
func TestLayoutInjectsGlancePalette(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	if !strings.Contains(html, "tailwind.config") {
		t.Errorf("la config Tailwind (palette) n'est pas injectée dans le head")
	}
	for _, token := range []string{"g-accent", "g-bg", "g-widget"} {
		if !strings.Contains(html, token) {
			t.Errorf("couleur %q absente de la palette Tailwind", token)
		}
	}
	for _, hex := range []string{"#002b36", "#b58900", "#93a1a1"} {
		if !strings.Contains(html, hex) {
			t.Errorf("couleur Solarized %q absente de la palette", hex)
		}
	}
}

// TestLayoutBodyStyleSheet : le style du body doit être dans un élément
// <style> réel (gomponents html.Style est un ATTRIBUT style="...", pas un
// élément — bug vu en prod : body restait blanc).
func TestLayoutBodyStyleSheet(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	if !strings.Contains(html, "<style>") {
		t.Fatalf("aucun élément <style> dans le head")
	}
	start := strings.Index(html, "<style>")
	end := strings.Index(html[start:], "</style>")
	if end < 0 {
		t.Fatalf("élément <style> non fermé")
	}
	css := html[start : start+end]
	if !strings.Contains(css, "background-color: #002b36") {
		t.Errorf("le fond Solarized (#002b36) n'est pas dans la feuille <style>")
	}
	// Focus visible partout (accessibilité)
	if !strings.Contains(css, ":focus-visible") {
		t.Errorf("la feuille <style> ne définit pas de focus-visible")
	}
	// L'attribut fantôme style="body{...} ne doit plus exister
	if strings.Contains(html, `style="body{`) {
		t.Errorf("attribut style fantôme toujours présent dans le head")
	}
}

func TestPageAccueilNoNestedAnchors(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}

	noNestedAnchors(t, renderedPage(t, PageAccueil("/", kbs, "")))
}

func TestOtherPagesNoNestedAnchors(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}

	noNestedAnchors(t, renderedPage(t, PageAccueil("/", nil, "")))                   // état vide
	noNestedAnchors(t, renderedPage(t, PageKBDetail("/kb/detail", kbs[0], kbs[1:]))) // détail
	noNestedAnchors(t, renderedPage(t, PageKBGraph("/kb/graph", kbs)))               // graphe
}

// TestPageAccueilDesignMinimal : rayons arrondis (rounded-xl), tags en pills,
// aucun → décoratif, tags pointant vers le filtre /?tag=xxx.
func TestPageAccueilDesignMinimal(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB06644754", "cc_flow")}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, "rounded-xl") {
		t.Errorf("les cartes doivent utiliser le rayon rounded-xl")
	}
	if strings.Contains(html, "→") {
		t.Errorf("la flèche décorative → doit être retirée")
	}
	if !strings.Contains(html, "/?tag=cc_flow") {
		t.Errorf("les tags doivent pointer vers le filtre /?tag=xxx")
	}
	if !strings.Contains(html, "rounded-full") {
		t.Errorf("les tags doivent être des pills (rounded-full)")
	}
}

// TestPageAccueilFiltreTag : le tag actif est affiché avec un moyen de le retirer.
func TestPageAccueilFiltreTag(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB22547", "quantite")}
	html := renderedPage(t, PageAccueil("/", kbs, "quantite"))

	if !strings.Contains(html, "Filtré par") {
		t.Errorf("le tag de filtre actif n'est pas affiché")
	}
	if !strings.Contains(html, "quantite") {
		t.Errorf("le nom du tag filtré n'est pas affiché")
	}
}

// TestExcerptGuard : garde-fou sur le helper d'extrait (sécurité runes).
func TestExcerptGuard(t *testing.T) {
	if got := excerpt("abcdefghij", 5); got != "abcde..." {
		t.Errorf("excerpt() = %q", got)
	}
	if got := excerpt("abc", 10); got != "abc" {
		t.Errorf("excerpt() court = %q", got)
	}
}
