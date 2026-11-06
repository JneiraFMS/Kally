package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"

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

// TestLayoutInjectsGlancePalette : le thème est chargé depuis la feuille
// locale /assets/tailwind.css (générée depuis tailwind.config.js qui porte
// la palette obsidienne/indigo) : plus de CDN JS (offline, pas de FOUC).
func TestLayoutInjectsGlancePalette(t *testing.T) {
	html := renderedPage(t, layout("T", "/", nil))

	if strings.Contains(html, "cdn.tailwindcss.com") {
		t.Errorf("le CDN Tailwind est encore référencé (mode dev, FOUC, offline impossible)")
	}
	if !strings.Contains(html, `<link rel="stylesheet" href="/assets/tailwind.css">`) {
		t.Errorf("la feuille locale /assets/tailwind.css n'est pas chargée:\n%s", html)
	}
	// La palette inline (runtime CDN) ne doit plus être injectée : la
	// source de vérité est tailwind.config.js, testée par palette_test.
	if strings.Contains(html, "tailwind.config") {
		t.Errorf("config Tailwind JS runtime encore injectée")
	}
	// Style du body consommant les tokens (aucun hex en dur)
	if !strings.Contains(html, "background-color: var(--g-bg)") {
		t.Errorf("couleur de fond du body non liée aux tokens")
	}
	// Le :root déclare le fond du thème sombre (source : palette.go)
	if !strings.Contains(html, paletteDark["g-bg"]) {
		t.Errorf("hex de fond du thème sombre (%s) absent du style body", paletteDark["g-bg"])
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
	// La feuille inline porte les variables de thème (:root dark =
	// valeur de référence, source palette.go) et le body les consomme.
	if !strings.Contains(css, "--g-bg:"+paletteDark["g-bg"]) {
		t.Errorf("le fond du thème sombre (%s) n'est pas déclaré dans les var du thème", paletteDark["g-bg"])
	}
	if !strings.Contains(css, "background-color: var(--g-bg)") {
		t.Errorf("le body ne consomme pas var(--g-bg)")
	}
	// Focus visible partout (accessibilité)
	if !strings.Contains(css, ":focus-visible") {
		t.Errorf("la feuille <style> ne définit pas de focus-visible")
	}
	if !strings.Contains(css, "scrollbar-gutter: stable") {
		t.Errorf("la largeur du viewport change selon la présence de scrollbar")
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

// TestPageFramesAlignWithNavbarLogo : la navbar et la zone de contenu
// principale partagent le même cadre max-w-5xl + gutter px-6, quelle que
// soit la page. Les contenus plus étroits restent contraints à droite, mais
// ne sont plus centrés sous un logo décalé.
func TestPageFramesAlignWithNavbarLogo(t *testing.T) {
	k := mustCardKB(t, "KB-0001", "demo")
	pages := []struct {
		name string
		node gomponents.Node
	}{
		{name: "accueil", node: PageAccueil("/", []kb.KB{k}, "")},
		{name: "création", node: PageKBCreate("/kb/new")},
		{name: "édition", node: PageKBEdit("/kb/edit", k)},
		{name: "détail", node: PageKBDetail("/kb/detail", dto.KBDetail{KB: k})},
		{name: "graphe", node: PageKBGraph("/kb/graph", []kb.KB{k})},
	}

	for _, tt := range pages {
		t.Run(tt.name, func(t *testing.T) {
			html := renderedPage(t, tt.node)
			if strings.Count(html, `max-w-5xl mx-auto px-6`) < 2 {
				t.Errorf("navbar et contenu ne partagent pas le cadre max-w-5xl/px-6:\n%s", html)
			}
		})
	}
}

func TestOtherPagesNoNestedAnchors(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB06644754", "cc_flow"),
		mustCardKB(t, "KB22547", "cc_flow", "quantite"),
	}

	noNestedAnchors(t, renderedPage(t, PageAccueil("/", nil, "")))                                                                  // état vide
	noNestedAnchors(t, renderedPage(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: kbs[0], Related: kbs[1:], Neighbors: kbs[1:]}))) // détail
	noNestedAnchors(t, renderedPage(t, PageKBGraph("/kb/graph", kbs)))                                                              // graphe
}

// TestPageAccueilDesignMinimal : rayons arrondis (rounded-xl), tags mono
// discrets (sans pilule épaisse), tag pointant vers /?tag=xxx — le lien
// étendu (toute la carte) est le seul déclencheur de navigation.
func TestPageAccueilDesignMinimal(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB06644754", "cc_flow")}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, "rounded-xl") {
		t.Errorf("les cartes doivent utiliser le rayon rounded-xl")
	}
	if strings.Count(html, "rounded-full") > 0 {
		t.Errorf("pas de pilule rounded-full : métadonnées en texte nu (zéro-pill)")
	}
	if !strings.Contains(html, "/?tag=cc_flow") {
		t.Errorf("les tags doivent pointer vers le filtre /?tag=xxx")
	}
	if !strings.Contains(html, "font-mono") {
		t.Errorf("identifiants/tags doivent être en mono")
	}
}

// TestPageAccueilToolbar : la barre d'outils de l'accueil est un en-tête de
// section clair — titre « Bases de connaissances », sous-titre explicatif,
// et bouton secondaire bordé (mode libre) ; le CTA primaire vit dans la
// navbar, il n'est pas dupliqué ici.
func TestPageAccueilToolbar(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB06644754", "cc_flow")}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, "Bases de connaissances") {
		t.Errorf("titre « Bases de connaissances » absent:\n%s", html)
	}
	if !strings.Contains(html, `href="/kb/new/quick"`) {
		t.Errorf("bouton « Nouvelle KB libre » sans lien")
	}
	if !strings.Contains(html, "Nouvelle KB libre") {
		t.Errorf("libellé « Nouvelle KB libre » absent")
	}
}

// TestPageAccueilFiltreTagSegmenté : les tags de filtrage sont des boutons
// segmentés ; le tag actif est marqué (plein accent + aria-pressed) et
// pointe vers l'URL SANS lui (retrait) ; les inactifs pointent vers
// l'URL avec eux (ajout) ; plus de bloc « Filtré par ».
func TestPageAccueilFiltreTagSegmenté(t *testing.T) {
	kbs := []kb.KB{
		mustCardKB(t, "KB22547", "quantite"),
		mustCardKB(t, "KB06644754", "quantite", "sap"),
	}
	html := renderedPage(t, PageAccueil("/", kbs, "quantite"))

	if strings.Contains(html, "Filtré par") {
		t.Errorf("bloc « Filtré par » encore présent (légende : boutons segmentés)")
	}
	if !strings.Contains(html, "aria-pressed=\"true\"") {
		t.Errorf("tag actif sans aria-pressed : l'état n'est pas accessible")
	}
	if !strings.Contains(html, `href="/"`) {
		t.Errorf("retrait du dernier tag actif doit retourner à l'accueil non filtré")
	}
	if !strings.Contains(html, `/?tag=quantite&amp;tag=sap`) {
		t.Errorf("tag inactif « sap » sans href d'ajout (URL combinée, &amp; échappé dans l'attribut)")
	}
	if !strings.Contains(html, "quantite") || !strings.Contains(html, "sap") {
		t.Errorf("libellés de tags absents des segments")
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

// TestSearchText : la clé de recherche fuzzy d'une KB contient l'ID, la
// première ligne de l'erreur et les tags, le tout en minuscules (les lignes
// suivantes de l'erreur sont exclues : on cherche sur ce qui est visible).
func TestSearchText(t *testing.T) {
	k, err := kb.NewKB("KB06644754", "Erreur de\nquantite négative", "notes", "résolution", "CC_Flow")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}

	got := searchText(k)
	for _, want := range []string{"kb06644754", "erreur de", "cc_flow"} {
		if !strings.Contains(got, want) {
			t.Errorf("searchText() = %q : %q manquant", got, want)
		}
	}
	if strings.Contains(got, "CC_Flow") || strings.Contains(got, "KB066") {
		t.Errorf("searchText() = %q : doit être entièrement en minuscules", got)
	}
	if strings.Contains(got, "négative") {
		t.Errorf("searchText() = %q : seule la première ligne de l'erreur doit matcher", got)
	}
}

// TestPageAccueilSearchBar : la barre de recherche fuzzy est rendue avec son
// script client, et chaque carte porte sa clé data-search (filtrage sans
// aucun aller-retour réseau).
func TestPageAccueilSearchBar(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB06644754", "cc_flow")}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, `id="kb-search"`) {
		t.Errorf("le champ de recherche id=kb-search est absent")
	}
	if !strings.Contains(html, `autocomplete="off"`) {
		t.Errorf("le champ de recherche doit désactiver l'autocomplete navigateur")
	}
	if !strings.Contains(html, `data-search="kb06644754 erreur démo cc_flow"`) {
		t.Errorf("la carte n'embarque pas sa clé data-search :\n%s", html)
	}
	if !strings.Contains(html, "fnFuzzyScore") {
		t.Errorf("le scoreur fuzzy client (fnFuzzyScore) n'est pas injecté")
	}
	if !strings.Contains(html, "id=\"kb-no-results\"") {
		t.Errorf("l'état « aucun résultat » est absent")
	}
}

// TestPageAccueilSearchMarkupGeneralized : le markup de la barre d'accueil
// est déclaratif (attributs data-fuzzy-*) — le moteur fuzzy est partagé avec
// le journal et toute future page, sans id figé dans le script.
func TestPageAccueilSearchMarkupGeneralized(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB06644754", "cc_flow")}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	for _, want := range []string{
		`data-fuzzy`,
		`data-fuzzy-count="kb-search-count"`,
		`data-fuzzy-no-results="kb-no-results"`,
		`data-fuzzy-order`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("markup de recherche généralisé sans %q", want)
		}
	}
}

// TestPageAccueilSearchBarVide : sans KB (états vides), ni barre ni script
// ne sont rendus — recherche = progressive enhancement inutile là.
func TestPageAccueilSearchBarVide(t *testing.T) {
	for _, tt := range []struct{ name, filtre string }{
		{"aucune KB", ""},
		{"tag sans résultat", "quantite"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var kbs []kb.KB
			html := renderedPage(t, PageAccueil("/", kbs, tt.filtre))

			if strings.Contains(html, "kb-search") || strings.Contains(html, "fnFuzzyScore") {
				t.Errorf("recherche rendue alors qu'aucune carte n'existe")
			}
		})
	}
}
