package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"
)

// sampleDetail construit un dto.KBDetail réaliste pour les tests de fragment.
func sampleDetail(t *testing.T, resolution string) dto.KBDetail {
	t.Helper()
	k := sampleKB(t, "KB06644754", "mm", true)
	k.Resolution = resolution
	return dto.KBDetail{
		KB:       k,
		KnownIDs: []string{k.ID},
	}
}

// TestPanelFragmentStructure : le fragment du tiroir est un node autonome
// (id="kb-panel"), sans layout (pas de doctype ni <nav>) — le shell du
// tiroir fournit la structure de page. Il porte l'en-tête mono ID · module
// · date, les sections non vides, le bandeau T-codes et le bloc copier.
func TestPanelFragmentStructure(t *testing.T) {
	html := renderedPage(t, PageKBDetailPanel(sampleDetail(t, "attendre la facture puis relancer /nVA01")))

	if !strings.Contains(html, `id="kb-panel"`) {
		t.Errorf("fragment sans conteneur kb-panel:\n%s", html)
	}
	if strings.Contains(strings.ToLower(html), "<!doctype") || strings.Contains(html, "<nav") {
		t.Errorf("fragment rendu comme une page complète (layout appelé)")
	}
	if !strings.Contains(html, "KB06644754") {
		t.Errorf("ID de la fiche absent du fragment")
	}
	if !strings.Contains(html, "SAP MM") {
		t.Errorf("module SAP absent de l'en-tête du fragment")
	}
	if !strings.Contains(html, "2026-11-04") {
		t.Errorf("date de modification absente du fragment")
	}
	// Sections rendues (non vides uniquement)
	for _, want := range []string{"Erreur", "Notes", "Résolution"} {
		if !strings.Contains(html, want) {
			t.Errorf("section %q absente du fragment", want)
		}
	}
	// Bandeau T-codes dérivé des notes
	if !strings.Contains(html, "ME22N") {
		t.Errorf("T-code dérivé absent du bandeau du fragment")
	}
}

// TestPanelFragmentAdaptative : une fiche libre sans Erreur/Résolution ne
// rend que ses sections existantes — jamais de « — vide — » décoratif.
func TestPanelFragmentAdaptative(t *testing.T) {
	k, err := kb.NewKB("KB-LIBRE", "note unique sans structure", "du markdown libre", "", "notes")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	html := renderedPage(t, PageKBDetailPanel(dto.KBDetail{KB: k, KnownIDs: []string{k.ID}}))

	if strings.Contains(html, "Résolution") {
		t.Errorf("section Résolution rendue alors que le champ est vide")
	}
	if strings.Contains(html, "— vide —") {
		t.Errorf("mention « — vide — » rendue (la forme doit suivre le contenu)")
	}
	if !strings.Contains(html, "du markdown libre") {
		t.Errorf("contenu libre absent du fragment")
	}
}

// TestPanelCopyBlock : le fragment porte le bloc « Copier la solution » —
// bouton (data-copy-target), textarea cachée source du presse-papiers,
// libellé de succès (data-copy-done) et script presse-papiers avec
// feedback 1500 ms. Le fallback natif (textarea + execCommand) couvre le
// presse-papiers non disponible.
func TestPanelCopyBlock(t *testing.T) {
	html := renderedPage(t, PageKBDetailPanel(sampleDetail(t, "attendre la facture")))

	if !strings.Contains(html, "Copier la solution") {
		t.Errorf("libellé du bloc copier absent")
	}
	if !strings.Contains(html, "data-copy-target") {
		t.Errorf("bouton copier sans data-copy-target")
	}
	if !strings.Contains(html, "<textarea") {
		t.Errorf("source presse-papiers (textarea) absente")
	}
	if !strings.Contains(html, "data-copy-done") {
		t.Errorf("libellé du succès (data-copy-done) absent")
	}
	if !strings.Contains(html, "1500") {
		t.Errorf("durée du feedback (1500 ms) absente du script copier")
	}
}

// TestSolutionText : le texte copié est prêt-à-coller (ServiceNow) :
// en-tête KB + module, symptôme, contexte, étapes, lien ServiceNow.
func TestSolutionText(t *testing.T) {
	k := sampleKB(t, "KB06644754", "mm", true)
	k.Resolution = "relancer la facture"

	got := solutionText(k)
	for _, want := range []string{
		"KB06644754",
		"SAP MM",
		"prix modifié interdit",
		"relancer la facture",
		serviceNowURL("KB06644754"),
	} {
		if !strings.Contains(got, want) {
			t.Errorf("solutionText() sans %q:\n%s", want, got)
		}
	}
}

// TestDrawerShell : le shell du tiroir vit sur l'accueil (pages avec
// cartes) : dialogue accessible (role/aria-modal/aria-label), overlay
// cliquable, bouton fermer, corps où le fragment s'injecte. Fermeture par
// défaut : pas de classe open au rendu.
func TestDrawerShell(t *testing.T) {
	kbs := []kb.KB{sampleKB(t, "KB06644754", "mm", true)}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	for _, want := range []string{
		`id="kb-drawer"`,
		`role="dialog"`,
		`aria-modal="true"`,
		`aria-label="Détail de la KB"`,
		`id="kb-drawer-body"`,
		"kb-drawer-overlay",
		"kb-drawer-close",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("shell du tiroir sans %s:\n%.60s", want, html)
		}
	}
	if strings.Contains(html, `kb-drawer open`) || strings.Contains(html, `class="open"`) {
		t.Errorf("tiroir rendu ouvert (fermeture par défaut attendue)")
	}
}

// TestDrawerScript : le script grz-drawer fetch le fragment /kb/panel,
// ouvre/ferme le tiroir (overlay, Échap), conserve une navigation de
// secours native vers /kb/detail si le fetch échoue.
func TestDrawerScript(t *testing.T) {
	kbs := []kb.KB{sampleKB(t, "KB06644754", "mm", true)}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	for _, want := range []string{
		"grz-drawer",
		`fetch("/kb/panel`,
		"kb-drawer-body",
		`location.href = anchor.getAttribute("href")`,
		`"Escape"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("script tiroir sans %s", want)
		}
	}
}

// TestDrawerAnchors : les ancres des cartes et des lignes du tableau
// portent data-kb-drawer (id de la KB) ET le href natif — le tiroir est
// une amélioration progressive : sans JS la navigation détail fonctionne.
func TestDrawerAnchors(t *testing.T) {
	kbs := []kb.KB{sampleKB(t, "KB06644754", "mm", true)}
	html := renderedPage(t, PageAccueil("/", kbs, ""))

	if !strings.Contains(html, `data-kb-drawer="KB06644754"`) {
		t.Errorf("ancre de carte sans data-kb-drawer")
	}
	// href natif conservé sur l'ancre étendue
	if !strings.Contains(html, `href="/kb/detail?id=KB06644754"`) {
		t.Errorf("href natif perdu sur l'ancre du tiroir")
	}
}

// TestDetailPageCopyBlock : la page détail historique porte aussi le bloc
// copier + son script (même comportement que le tiroir).
func TestDetailPageCopyBlock(t *testing.T) {
	k := sampleKB(t, "KB06644754", "mm", true)
	k.Resolution = "relancer la facture"
	html := renderedPage(t, PageKBDetail("/kb/detail", dto.KBDetail{KB: k, KnownIDs: []string{k.ID}}))

	if !strings.Contains(html, "Copier la solution") {
		t.Errorf("bloc copier absent de la page détail")
	}
	if !strings.Contains(html, "1500") {
		t.Errorf("script copier absent de la page détail")
	}
}
