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

func TestPageAccueilNoNestedAnchors(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}

	noNestedAnchors(t, renderedPage(t, PageAccueil("/", kbs)))
}

func TestOtherPagesNoNestedAnchors(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}

	noNestedAnchors(t, renderedPage(t, PageAccueil("/", nil)))                       // état vide
	noNestedAnchors(t, renderedPage(t, PageKBList("/kb", kbs)))                      // liste
	noNestedAnchors(t, renderedPage(t, PageKBDetail("/kb/detail", kbs[0], kbs[1:]))) // détail
	noNestedAnchors(t, renderedPage(t, PageKBGraph("/kb/graph", kbs)))               // graphe
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
