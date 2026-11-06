package pages

import (
	"strings"
	"testing"
	"time"

	kb "grz/internal/domain/kb"
)

// sampleKB construit une KB réaliste : erreur multi-lignes (titre + trace),
// un T-code dans les notes, tags module et projet, date de modification.
func sampleKB(t *testing.T, id, tagModule string, modified bool) kb.KB {
	t.Helper()
	k, err := kb.NewKB(id,
		"Message 15421 : prix modifié interdit\nquantite négative refusée à la validation",
		"pousser /nME22N puis vérifier la commande", "résolution", "cc_flow", tagModule)
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	if modified {
		k.Modified = time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC)
	}
	return k
}

// TestKbCardStructure : la carte suit le canon « zéro-pilule » —
// en-tête mono non encadré `ID · SAP MM · date` (date en tabular-nums à
// droite), titre line-clamp-2, extrait, pied avec tag cliquable mono et
// premier T-code. Le lien étendu reste le seul moyen de
// navigation (pas de <a> imbriqué).
func TestKbCardStructure(t *testing.T) {
	html := renderedPage(t, PageAccueil("/", []kb.KB{sampleKB(t, "KB06644754", "mm", true)}, ""))

	if !strings.Contains(html, ">KB06644754</") {
		t.Errorf("ID de la carte absent:\n%s", html)
	}
	if !strings.Contains(html, "text-g-accent-text") {
		t.Errorf("ID de la carte sans la couleur accent")
	}
	if !strings.Contains(html, "·") {
		t.Errorf("séparateur « · » absent de l'en-tête de carte")
	}
	if !strings.Contains(html, "SAP MM") {
		t.Errorf("module SAP dérivé du tag absent")
	}
	if !strings.Contains(html, "2026-11-04") {
		t.Errorf("date de dernière modification absente")
	}
	if !strings.Contains(html, "tabular-nums") {
		t.Errorf("date sans tabular-nums (alignement vertical dégradé)")
	}
	for _, want := range []string{"prix modifié interdit", "quantite négative refusée"} {
		if !strings.Contains(html, want) {
			t.Errorf("contenu de carte sans %q (titre = 1re ligne, extrait = suite)", want)
		}
	}
	if !strings.Contains(html, `href="/?tag=cc_flow"`) {
		t.Errorf("tag principal non cliquable vers le filtre")
	}
	if !strings.Contains(html, "ME22N") {
		t.Errorf("T-code SAP dérivé absent du pied de carte")
	}
	if !strings.Contains(html, `href="/kb/detail?id=KB06644754"`) {
		t.Errorf("lien étendu vers le détail absent")
	}
	if !strings.Contains(html, "absolute inset-0") {
		t.Errorf("technique du lien étendu absente (overlay)")
	}
}

// TestKbCardSansDate : sans mtime (tests, KB fraîchement parsée sans
// stat), la date ne s'affiche pas — pas de date inventée.
func TestKbCardSansDate(t *testing.T) {
	html := renderedPage(t, PageAccueil("/", []kb.KB{sampleKB(t, "KB06644754", "mm", false)}, ""))

	if strings.Contains(html, "2026-11-04") {
		t.Errorf("date affichée alors que Modified est zéro")
	}
	if !strings.Contains(html, ">KB06644754</") {
		t.Errorf("ID de la carte absent")
	}
}

// TestKbCardSansModule : sans tag module, le libellé « SAP × » est absent,
// le module dérivé ne s'invente pas.
func TestKbCardSansModule(t *testing.T) {
	html := renderedPage(t, PageAccueil("/", []kb.KB{sampleKB(t, "KB06644754", "cc_ancien", false)}, ""))

	if strings.Contains(html, "SAP ") {
		t.Errorf("libellé SAP affiché alors qu'aucun tag module n'existe")
	}
}

// TestKbTableDenseView : la vue tableau haute densité est rendue cachée
// par défaut (les cartes restent la vue par défaut), avec en-têtes
// discrets, une ligne par KB (h-9, hover), ID en mono cliquable, module,
// titre tronqué, tags, T-codes, date. Le fuzzy compte UNE fois par KB :
// les deux vues vivent dans des conteneurs data-fuzzy-view distincts.
func TestKbTableDenseView(t *testing.T) {
	kbs := []kb.KB{
		sampleKB(t, "KB06644754", "mm", true),
		sampleKB(t, "KB22547", "sd", false),
	}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, `id="kb-cards"`) || !strings.Contains(html, `data-fuzzy-view="grid"`) {
		t.Errorf("conteneur cartes sans id/data-fuzzy-view=grid")
	}
	if !strings.Contains(html, `id="kb-table"`) || !strings.Contains(html, `data-fuzzy-view="table"`) {
		t.Errorf("conteneur tableau sans id/data-fuzzy-view=table")
	}
	// À l'affichage, le conteneur reste aligné (mx-auto) mais plus large
	// que la grille : le tableau dense tient ses 6 colonnes sans garnir les
	// lignes. `hidden` est seulement togglé par la bascule.
	if !strings.Contains(html, `id="kb-table" class="hidden max-w-7xl mx-auto px-6 pb-8"`) {
		t.Errorf("conteneur tableau sans largeur/cadre: conteneur moins large que depuis sa création")
	}
	// Le tableau dense vit dans une surface carte arrondie et bordée.
	if !strings.Contains(html, `rounded-xl border border-g-border bg-g-widget`) {
		t.Errorf("tableau sans surface carte (arrondie + bordée)")
	}
	// Ligne d'entête distincte (fond bloc) et cellules qui respirent.
	for _, want := range []string{
		`<thead><tr class="bg-g-block border-b border-g-border">`,
		`class="px-4 py-2.5`,
		`px-4 font-mono text-xs`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("style du tableau sans %q : markup sans la classe", want)
		}
	}
	if !strings.Contains(html, "hidden") {
		t.Errorf("vue tableau visible par défaut (doit être cachée)")
	}

	for _, head := range []string{"Module", "Tag", "T-codes"} {
		if !strings.Contains(html, head) {
			t.Errorf("en-tête de colonne absente : %q", head)
		}
	}

	if got := strings.Count(html, "<tr"); got != 3 { // 1 tête + 2 lignes
		t.Errorf("nombre de <tr> = %d, want 3 (1 entête + 2 lignes)", got)
	}
	// ID cliquable en mono dans la cellule
	if !strings.Contains(html, `href="/kb/detail?id=KB06644754"`) {
		t.Errorf("Id de la ligne non cliquable")
	}
}

// TestKbTableSansKB : sans carte, ni tableau ni script de bascule. La
// barre grz-local mentionne l'ancre d'injection du tableau côté script
// (élément légitime de la bibliothèque locale) : on porte la vérification
// sur l'élément rendu, pas sur la sous-chaîne.
func TestKbTableSansKB(t *testing.T) {
	var kbs []kb.KB
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if strings.Contains(html, `id="kb-table"`) || strings.Contains(html, "grz-view") {
		t.Errorf("vue tableau/script de bascule rendus alors qu'aucune carte n'existe")
	}
}

// TestPageAccueilViewToggleScript : la bascule cartes/tableau est(script)
// côté client : deux boutons icônes (aria-pressed portant l'état), état
// persisté en localStorage grz-view, cartes visibles par défaut.
func TestPageAccueilViewToggleScript(t *testing.T) {
	kbs := []kb.KB{sampleKB(t, "KB06644754", "mm", true)}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, `data-view-target="grid"`) || !strings.Contains(html, `data-view-target="table"`) {
		t.Errorf("boutons de bascule sans data-view-target")
	}
	if !strings.Contains(html, "grz-view") {
		t.Errorf("script de bascule sans clé localStorage grz-view")
	}
	for _, want := range []string{"aria-pressed", "Vue cartes", "Vue tableau"} {
		if !strings.Contains(html, want) {
			t.Errorf("bascule sans %q", want)
		}
	}
}

// TestFuzzyCountsPerVisibleView : le moteur fuzzy ne compte chaque élément
// que si son conteneur data-fuzzy-view est visible — 11 cartes + 11 lignes
// s'affichent « 11 / 11 », jamais « 22 / 11 ».
func TestFuzzyCountsPerVisibleView(t *testing.T) {
	if !strings.Contains(fuzzyScript, "data-fuzzy-view") {
		t.Errorf("moteur fuzzy sans contrat de conteneur data-fuzzy-view")
	}
	if !strings.Contains(fuzzyScript, "closest") {
		t.Errorf("moteur fuzzy sans rattachement au conteneur visible (closest)")
	}
}
