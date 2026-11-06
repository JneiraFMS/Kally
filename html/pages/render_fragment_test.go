package pages

import (
	"strings"
	"testing"
)

// TestRenderMarkdownHTMLRendersSafeFragment : le fragment markdown consommé
// par le script grz-local est sûr (HTML brut ignoré) et rendu dans .md.
func TestRenderMarkdownHTMLRendersSafeFragment(t *testing.T) {
	out := RenderMarkdownHTML("<script>evil()</script>\n\n**gras**", nil)
	if !strings.Contains(out, `<div class="md">`) {
		t.Errorf("RenderMarkdownHTML() sans conteneur .md : %q", out)
	}
	if !strings.Contains(out, "raw HTML omitted") && strings.Contains(out, "<script>") {
		t.Errorf("RenderMarkdownHTML() HTML brut exécutable : %q", out)
	}
	if !strings.Contains(out, "<strong>gras</strong>") {
		t.Errorf("RenderMarkdownHTML() enrichissements markdown perdus : %q", out)
	}
}

// TestRenderMarkdownHTMLResolvesWikiLinksAgainstKnownIDs : un fichier de la
// bibliothèque locale cite le serveur : avec known fourni (même vide), une
// cible absente est un lien mort ; sans known (nil), le résolveur est
// inconnu et tous les liens sont rendus existants.
func TestRenderMarkdownHTMLResolvesWikiLinksAgainstKnownIDs(t *testing.T) {
	live := RenderMarkdownHTML("[[KB-2]]", nil)
	if !strings.Contains(live, `href="/kb/detail?id=KB-2"`) {
		t.Errorf("résolveur inconnu : %q", live)
	}

	dead := RenderMarkdownHTML("[[KB-X]]", []string{"KB-2"})
	if !strings.Contains(dead, "wikilink-missing") || !strings.Contains(dead, `href="/kb/new?id=KB-X"`) {
		t.Errorf("lien mort attendu : %q", dead)
	}

	known := RenderMarkdownHTML("[[KB-2]]", []string{"KB-2"})
	if !strings.Contains(known, `href="/kb/detail?id=KB-2"`) {
		t.Errorf("lien vivant attendu : %q", known)
	}
}
