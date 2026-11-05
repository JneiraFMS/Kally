package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"
)

func wikiDetail(t *testing.T, notes, resolution string, known []string) string {
	t.Helper()
	k, err := kb.NewKB("KB-0001", "erreur", notes, resolution)
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	return renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k, KnownIDs: known}))
}

// TestWikiLinkRendersExistingLink : une cible connue devient un lien interne
// vers la fiche, avec la classe wikilink (style dédié).
func TestWikiLinkRendersExistingLink(t *testing.T) {
	html := wikiDetail(t, "voir [[KB-0002]] pour la suite", "", []string{"KB-0001", "KB-0002"})

	if !strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("lien wiki existant sans href vers la fiche:\n%s", html)
	}
	if !strings.Contains(html, `class="wikilink"`) {
		t.Errorf("lien wiki sans classe wikilink:\n%s", html)
	}
	if strings.Contains(html, `class="wikilink wikilink-missing"`) {
		t.Errorf("une cible connue est marquée manquante:\n%s", html)
	}
	if !strings.Contains(html, ">KB-0002</a>") {
		t.Errorf("le libellé du lien wiki doit être la cible:\n%s", html)
	}
}

// TestWikiLinkRendersMissingLink : une cible absente devient un lien mort
// vers le formulaire de création prérempli, distinct visuellement.
func TestWikiLinkRendersMissingLink(t *testing.T) {
	html := wikiDetail(t, "voir [[KB-9999]]", "", []string{"KB-0001"})

	if !strings.Contains(html, `href="/kb/new?id=KB-9999"`) {
		t.Errorf("lien mort sans href vers la création préremplie:\n%s", html)
	}
	if !strings.Contains(html, `class="wikilink wikilink-missing"`) {
		t.Errorf("lien mort sans classe wikilink-missing:\n%s", html)
	}
	if !strings.Contains(html, `title="Créer la KB KB-9999"`) {
		t.Errorf("lien mort sans indication de création:\n%s", html)
	}
}

// TestWikiLinkWithoutKnownIDsTreatsLinksAsExisting : sans résolveur (rendu
// markdown générique, journal), un lien reste un lien normal.
func TestWikiLinkWithoutKnownIDsTreatsLinksAsExisting(t *testing.T) {
	html := renderNode(t, mdToNode("voir [[KB-0002]]"))

	if !strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("lien wiki sans résolveur non rendu:\n%s", html)
	}
	if strings.Contains(html, `class="wikilink wikilink-missing"`) {
		t.Errorf("lien wiki sans résolveur marqué manquant:\n%s", html)
	}
}

// TestWikiLinkIgnoredInCode : un [[...]] dans du code (inline ou fencé)
// reste littéral, comme pour l'extraction des rétroliens.
func TestWikiLinkIgnoredInCode(t *testing.T) {
	html := wikiDetail(t,
		"inline `[[KB-0002]]`\n\n```\n[[KB-0003]]\n```\n",
		"", []string{"KB-0001", "KB-0002", "KB-0003"})

	if strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("lien wiki rendu dans du code inline:\n%s", html)
	}
	if strings.Contains(html, `href="/kb/detail?id=KB-0003"`) {
		t.Errorf("lien wiki rendu dans un bloc de code:\n%s", html)
	}
	if !strings.Contains(html, "[[KB-0002]]") || !strings.Contains(html, "[[KB-0003]]") {
		t.Errorf("le texte [[...]] du code doit rester littéral:\n%s", html)
	}
}

// TestWikiLinkInvalidTargetStaysLiteral : une cible hors grammaire n'est pas
// un lien (aucune injection possible).
func TestWikiLinkInvalidTargetStaysLiteral(t *testing.T) {
	html := wikiDetail(t, "[[pas un id]] et [[<script>alert(1)</script>]]", "", []string{"KB-0001"})

	if strings.Contains(html, `class="wikilink"`) {
		t.Errorf("cible invalide rendue comme lien wiki:\n%s", html)
	}
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Errorf("contenu de cible injecté sans échappement:\n%s", html)
	}
}

// TestWikiLinkRendersInResolution : la section Résolution est aussi liée
// (Notes et Résolution sont les deux sections markdown actives).
func TestWikiLinkRendersInResolution(t *testing.T) {
	html := wikiDetail(t, "", "corrigé par [[KB-0002]]", []string{"KB-0001", "KB-0002"})

	if !strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("lien wiki de la Résolution non rendu:\n%s", html)
	}
}

// TestWikiLinkNotRenderedInErreur : la section Erreur reste une trace brute
// (monospace, non liée) même si elle contient [[...]].
func TestWikiLinkNotRenderedInErreur(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "trace [[KB-0002]]", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{
		KB:       k,
		KnownIDs: []string{"KB-0001", "KB-0002"},
	}))

	if strings.Contains(html, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("la section Erreur ne doit pas produire de lien wiki:\n%s", html)
	}
	if !strings.Contains(html, "[[KB-0002]]") {
		t.Errorf("le texte de la section Erreur doit rester littéral:\n%s", html)
	}
}

// TestPageKBDetailShowsBacklinks : la fiche expose les KB qui la citent,
// avec un lien interne et un extrait de leur erreur.
func TestPageKBDetailShowsBacklinks(t *testing.T) {
	k, err := kb.NewKB("KB-0002", "erreur cible", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	source, err := kb.NewKB("KB-0001", "erreur source", "[[KB-0002]]", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{
		KB:        k,
		Backlinks: []kb.KB{source},
		KnownIDs:  []string{"KB-0001", "KB-0002"},
	}))

	if !strings.Contains(html, "Références liées") {
		t.Errorf("section « Références liées » absente:\n%s", html)
	}
	if !strings.Contains(html, `href="/kb/detail?id=KB-0001"`) {
		t.Errorf("rétrolien sans lien vers la KB source:\n%s", html)
	}
	if !strings.Contains(html, "erreur source") {
		t.Errorf("extrait de la KB source absent:\n%s", html)
	}
}

func TestPageKBDetailBacklinksEmpty(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k}))

	if !strings.Contains(html, "Aucune référence liée") {
		t.Errorf("état vide des rétroliens absent:\n%s", html)
	}
}

// TestPageKBGraphRendersExplicitLinkEdge : un lien [[...]] produit une arête
// distincte (kg-edge-link) et un tooltip dédié, même sans tag partagé.
func TestPageKBGraphRendersExplicitLinkEdge(t *testing.T) {
	a, err := kb.NewKB("KB-0001", "e", "[[KB-0002]]", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	b, err := kb.NewKB("KB-0002", "e", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderedPage(t, PageKBGraph("/kb/graph", []kb.KB{a, b}))

	if !strings.Contains(html, "kg-edge-link") {
		t.Errorf("arête de lien explicite sans classe kg-edge-link:\n%s", html)
	}
	if !strings.Contains(html, "Lien explicite") {
		t.Errorf("tooltip d'arête de lien explicite absent:\n%s", html)
	}
	if !strings.Contains(html, `data-a="KB-0001"`) || !strings.Contains(html, `data-b="KB-0002"`) {
		t.Errorf("arête de lien sans data-a/data-b:\n%s", html)
	}
}

// TestPageKBDetailLocalGraphShowsExplicitLinkNeighbor : un voisin relié
// uniquement par [[...]] apparaît dans le mini-graphe avec une arête liée.
func TestPageKBDetailLocalGraphShowsExplicitLinkNeighbor(t *testing.T) {
	k, err := kb.NewKB("KB-0001", "e", "[[KB-0002]]", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	neighbor, err := kb.NewKB("KB-0002", "voisin", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	html := renderNode(t, PageKBDetail("/kb/detail", dto.KBDetail{
		KB:        k,
		Neighbors: []kb.KB{neighbor},
		KnownIDs:  []string{"KB-0001", "KB-0002"},
	}))

	if !strings.Contains(html, `data-kg="local"`) {
		t.Fatalf("mini-graphe absent pour un voisin lié:\n%s", html)
	}
	if !strings.Contains(html, "kg-edge-link") {
		t.Errorf("arête de lien explicite absente du mini-graphe:\n%s", html)
	}
	if !strings.Contains(html, "Lien explicite") {
		t.Errorf("tooltip d'arête liée absent du mini-graphe:\n%s", html)
	}
}
