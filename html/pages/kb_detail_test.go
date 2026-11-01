package pages

import (
	"bytes"
	"strings"
	"testing"

	kb "grz/internal/domain/kb"

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
	k, err := kb.NewKB("KB-0001", "Division par zéro", "", "Vérifier le diviseur", "calcul")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	related := []kb.KB{}
	if k2, err := kb.NewKB("KB-0002", "Autre problème", "n", "r", "calcul"); err == nil {
		related = append(related, k2)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k, related))

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

	// Badge tag cliquable vers le filtre (migré sur /?tag=)
	if !strings.Contains(html, "/?tag=calcul") {
		t.Errorf("badge tag cliquable manquant")
	}

	// KB liée par tag : lien bidirectionnel vers KB-0002
	if !strings.Contains(html, "/kb/detail?id=KB-0002") {
		t.Errorf("lien vers KB liée manquant")
	}

	// Pas les anciens blocs gris mono
	if strings.Contains(html, "bg-gray-50") {
		t.Errorf("ancien style bg-gray-50 encore présent")
	}
}

// TestPageKBDetailRendersSectionsContent : la section Erreur est rendue en
// bloc monospace (traces de pile lues ligne à ligne), Notes/Résolution sont
// rendues comme du markdown (gras, listes, code), le HTML brut du contenu
// markdown n'est jamais exécuté.
func TestPageKBDetailRendersSectionsContent(t *testing.T) {
	k, err := kb.NewKB(
		"KB-0001",
		"panic: runtime error\n\n\tat module.go:42 in call()",
		"voir **les notes** et `une commande`",
		"## Résolution en markdown\n- étape un\n- étape deux",
	)
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k, nil))

	// Erreur : bloc <pre> monospace, texte brut (pas du markdown)
	if !strings.Contains(html, `<pre class="font-mono`) {
		t.Errorf("la section Erreur doit être un <pre> monospace:\n%s", html)
	}
	if !strings.Contains(html, "panic: runtime error") {
		t.Errorf("le texte de la section Erreur est perdu")
	}

	// Notes : markdown rendu
	if !strings.Contains(html, "<strong>les notes</strong>") {
		t.Errorf("le markdown des Notes n'est pas rendu (gras absent):\n%s", html)
	}
	if !strings.Contains(html, "<code>une commande</code>") {
		t.Errorf("le code inline des Notes n'est pas rendu")
	}

	// Résolution : markdown rendu (liste)
	if !strings.Contains(html, "<h2>Résolution en markdown</h2>") {
		t.Errorf("le titre markdown de la Résolution n'est pas rendu:\n%s", html)
	}
	if !strings.Contains(html, "<li>") {
		t.Errorf("la liste de la Résolution n'est pas rendue")
	}

	// Classe conteneur pour les styles markdown
	if !strings.Contains(html, `class="md"`) {
		t.Errorf("les sections markdown doivent porter la classe md")
	}

	// Conteneur de code stylé (bloc) pour les sections markdown
	if !strings.Contains(html, ".md pre") {
		t.Errorf("styles markdown manquants (.md pre)")
	}
}

// TestPageKBDetailMarkdownIsSafe : du HTML brut dans le markdown ne doit pas
// ressortir tel quel (aucune option Unsafe).
func TestPageKBDetailMarkdownIsSafe(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "notes <script>alert(1)</script>", "<b>gras</b>")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k, nil))

	if strings.Contains(html, "<script>alert(1)") {
		t.Errorf("le HTML brut du markdown est injecté tel quel:\n%s", html)
	}
	if strings.Contains(html, "<b>gras</b>") {
		t.Errorf("le HTML brut inline est injecté tel quel:\n%s", html)
	}
}

// TestPageKBDetailEmptySectionsStillMarked : les sections vides gardent la
// mention « — vide — » (avant/après le nouveau rendu). Ces mentions et le
// compte des KB ne doivent pas rester au niveau délavé g-subdue
// (contraste insuffisant — passe accessibilité Phase 3).
func TestPageKBDetailEmptySectionsStillMarked(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "Erreur seule", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k, nil))

	if strings.Count(html, "— vide —") != 2 {
		t.Errorf("sections vides non marquées « — vide — »:\n%s", html)
	}
	if strings.Contains(html, "text-g-subdue") {
		t.Errorf("mentions importantes encore en g-subdue (contraste):\n%s", html)
	}
	if !strings.Contains(html, `<i class="text-g-base"`) {
		t.Errorf("mention « vide » sans classe contrastée g-base:\n%s", html)
	}
}

// TestPageKBDetailHasDeleteAction : la barre d'actions expose une
// suppression en POST natif (fonctionne sans JS) avec confirmation htmx.
func TestPageKBDetailHasDeleteAction(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "n", "r")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", k, nil))

	if !strings.Contains(html, `action="/kb/delete"`) || !strings.Contains(html, `method="post"`) {
		t.Errorf("formulaire de suppression natif manquant:\n%s", html)
	}
	if !strings.Contains(html, `name="id" value="KB-0001"`) {
		t.Errorf("champ caché id du formulaire de suppression manquant")
	}
	if !strings.Contains(html, "hx-confirm") {
		t.Errorf("confirmation htmx de suppression manquante")
	}
	if !strings.Contains(html, ">Supprimer</button>") {
		t.Errorf("bouton Supprimer manquant")
	}
}
