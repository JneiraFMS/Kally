package pages

import (
	"bytes"
	"strings"
	"testing"

	"grz/internal/kb"

	"maragu.dev/gomponents"
)

// renderNode sérialise un composant gomponents en HTML.
func renderNode(t *testing.T, n gomponents.Node) string {
	t.Helper()
	var buf bytes.Buffer
	if err := n.Render(&buf); err != nil {
		t.Fatalf("Render() inattendu: %v", err)
	}
	return buf.String()
}

func TestPageKBDetailRendersNoteStyle(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "Division par zéro", "", "Vérifier le diviseur")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k))

	// Titre de la note en H1
	if !strings.Contains(html, ">KB-0001</h1>") {
		t.Errorf("le titre de note n'est pas un H1:\n%s", html)
	}

	// Chaque section est un H2 suivi de son texte
	for _, titre := range []string{"Erreur", "Notes", "Résolution"} {
		if !strings.Contains(html, "<h2") || !strings.Contains(html, ">"+titre+"</h2>") {
			t.Errorf("la section %q n'est pas un H2", titre)
		}
	}

	// Texte des sections remplies
	if !strings.Contains(html, "Division par zéro") || !strings.Contains(html, "Vérifier le diviseur") {
		t.Errorf("texte de section manquant")
	}

	// Section vide : mention "— vide —"
	if !strings.Contains(html, "— vide —") {
		t.Errorf("la section vide n'affiche pas la mention '— vide —'")
	}

	// Pas les anciens blocs gris mono
	if strings.Contains(html, "bg-gray-50") {
		t.Errorf("ancien style bg-gray-50 encore présent")
	}
}
