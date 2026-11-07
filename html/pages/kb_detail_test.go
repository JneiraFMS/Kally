package pages

import (
	"bytes"
	"strings"
	"testing"

	"grz/internal/application/dto"
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

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k, Related: related, Neighbors: related}))

	// Titre de la note : h1 (sémantique) contenant le lien vers l'article
	// ServiceNow de la KB, ouverture dans un nouvel onglet sécurisée.
	if !strings.Contains(html, `<h1 class="font-mono`) {
		t.Errorf("le titre de note n'est pas un H1:\n%s", html)
	}
	for _, want := range []string{
		`href="https://airbus.service-now.com/kb_view.do?sysparm_article=`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		">KB-0001</a></h1>",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("le titre ne pointe pas l'article ServiceNow (%q):\n%s", want, html)
		}
	}

	// Chaque section remplie est un H2 suivi som texte ; la section vide
	// (Notes ici) ne s'affiche pas du tout.
	for _, titre := range []string{"Erreur", "Résolution"} {
		if !strings.Contains(html, "<h2") || !strings.Contains(html, ">"+titre+"</h2>") {
			t.Errorf("la section %q n'est pas un H2", titre)
		}
	}

	// Texte des sections remplies
	if !strings.Contains(html, "Division par zéro") || !strings.Contains(html, "Vérifier le diviseur") {
		t.Errorf("texte de section manquant")
	}

	// La section vide n'existe plus dans le rendu : ni titre, ni Mention « — vide — ».
	if strings.Contains(html, "— vide —") {
		t.Errorf("section vide encore rendue (mention « — vide — »):\n%s", html)
	}

	// Badge tag cliquable vers le filtre (migré sur /?tag=)
	if !strings.Contains(html, "/?tag=calcul") {
		t.Errorf("badge tag cliquable manquant")
	}

	// KB liée par tag : navigation interne conservée + lien vers l'article
	// ServiceNow, compact (pas les classes d'un titre).
	if !strings.Contains(html, "/kb/detail?id=KB-0002") {
		t.Errorf("lien interne vers KB liée manquant:\n%s", html)
	}
	if !strings.Contains(html, `sysparm_article=KB-0002`) {
		t.Errorf("lien ServiceNow de la KB liée manquant:\n%s", html)
	}
	if strings.Contains(html, `text-xl font-semibold text-g-hi hover:underline">KB-0002`) {
		t.Errorf("la KB liée réutilise les classes du titre (style de liste écrasé):\n%s", html)
	}

	// Pas les anciens blocs gris mono
	if strings.Contains(html, "bg-gray-50") {
		t.Errorf("ancien style bg-gray-50 encore présent")
	}

	// Aucun débris de débogage jeté dans la note
	if strings.Contains(html, "sdfsdf") {
		t.Errorf("débris de débogage rendu dans la note:\n%s", html)
	}
}

// TestPageKBDetailHidesEmptySections : la forme suit le contenu — une KB
// créée en mode libre (seule la note est remplie) n'affiche pas les sections
// Erreur et Résolution vides.
func TestPageKBDetailHidesEmptySections(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "", "note libre", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if strings.Contains(html, ">Erreur</h2>") || strings.Contains(html, ">Résolution</h2>") {
		t.Errorf("sections vides rendues malgré leur absence de contenu:\n%s", html)
	}
	if !strings.Contains(html, ">Notes</h2>") {
		t.Errorf("section Notes absente du détail:\n%s", html)
	}
	if !strings.Contains(html, "note libre") {
		t.Errorf("contenu de la KB libre absent:\n%s", html)
	}
}

// TestPageKBDetailKeepsFilledSections : une section remplie apparaît ; une
// section vide reste masquée, même si une autre est prolongée.
func TestPageKBDetailKeepsFilledSections(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "trace d'erreur", "notes", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if !strings.Contains(html, ">Erreur</h2>") || !strings.Contains(html, ">Notes</h2>") {
		t.Errorf("sections remplies invisibles:\n%s", html)
	}
	if strings.Contains(html, ">Résolution</h2>") {
		t.Errorf("section Résolution vide rendue:\n%s", html)
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

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

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

// TestPageKBDetailErreurCodeSurface : la trace de la section Erreur utilise
// la surface de code (g-code) et un texte g-hi : g-text tomberait sous
// 4.5:1 sur ce fond plus clair.
func TestPageKBDetailErreurCodeSurface(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "panic: runtime error", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	start := strings.Index(html, `<pre class="font-mono`)
	if start < 0 {
		t.Fatalf("section Erreur sans <pre> monospace:\n%s", html)
	}
	end := strings.Index(html[start:], ">")
	tag := html[start : start+end+1]

	for _, want := range []string{"bg-g-code", "text-g-hi", "border-g-border"} {
		if !strings.Contains(tag, want) {
			t.Errorf("le <pre> de la section Erreur n'a pas %q :\n%s", want, tag)
		}
	}
	if strings.Contains(tag, "text-g-text") {
		t.Errorf("le <pre> de la section Erreur garde text-g-text (contraste insuffisant sur g-code) :\n%s", tag)
	}
}

// TestPageKBDetailMarkdownIsSafe : du HTML brut dans le markdown ne doit pas
// ressortir tel quel (aucune option Unsafe).
func TestPageKBDetailMarkdownIsSafe(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "notes <script>alert(1)</script>", "<b>gras</b>")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if strings.Contains(html, "<script>alert(1)") {
		t.Errorf("le HTML brut du markdown est injecté tel quel:\n%s", html)
	}
	if strings.Contains(html, "<b>gras</b>") {
		t.Errorf("le HTML brut inline est injecté tel quel:\n%s", html)
	}
}

// TestPageKBDetailEmptySectionsHidden : les sections vides disparaissent de
// la fiche (une KB créée en mode libre garde sa forme) ; en aucun cas une
// mention délavée n'apparaît (couvert au contaste par l'absence de rendu).
func TestPageKBDetailEmptySectionsHidden(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "Erreur seule", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if strings.Count(html, "— vide —") != 0 {
		t.Errorf("sections vides encore rendues (mention « — vide — »):\n%s", html)
	}
	if strings.Contains(html, ">Notes</h2>") || strings.Contains(html, ">Résolution</h2>") {
		t.Errorf("titres de sections vides rendus:\n%s", html)
	}
	if strings.Contains(html, "text-g-subdue") {
		t.Errorf("mentions importantes encore en g-subdue (contraste):\n%s", html)
	}
}

// TestPageKBDetailHasDeleteAction : la barre d'actions expose une
// suppression en POST natif (fonctionne sans JS) avec confirmation htmx.
func TestPageKBDetailHasDeleteAction(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "n", "r")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if !strings.Contains(html, `action="/kb/delete"`) || !strings.Contains(html, `method="post"`) {
		t.Errorf("formulaire de suppression natif manquant:\n%s", html)
	}
	if !strings.Contains(html, `name="id" value="KB-0001"`) {
		t.Errorf("champ caché id du formulaire de suppression manquant")
	}
	// Confirmation UNIQUE et native (dialogue du navigateur), valable avec
	// comme sans htmx — hx-confirm en doublon demandait deux confirmations
	// quand htmx est actif (le dialogue natif s'exécute au dispatch).
	if !strings.Contains(html, `onsubmit`) || !strings.Contains(html, "confirm") {
		t.Errorf("confirmation native de suppression manquante:\n%s", html)
	}
	if strings.Contains(html, "hx-confirm") {
		t.Errorf("double confirmation htmx indue (dialogue natif déjà posé)")
	}
	if !strings.Contains(html, ">Supprimer</button>") {
		t.Errorf("bouton Supprimer manquant")
	}
}
