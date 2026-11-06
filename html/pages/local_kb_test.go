package pages

import (
	"strings"
	"testing"

	kb "grz/internal/domain/kb"
)

// TestPageAccueilContainsLocalBar : la barre des articles locaux est rendue
// cachée (hidden) sur l'accueil, même sans KB partagée, la grille de cartes
// est toujours présente (les articles locaux s'invitent dedans), l'état
// vide porte sa bascule et le shell du tiroir est toujours rendu.
func TestPageAccueilContainsLocalBar(t *testing.T) {
	page := renderedPage(t, PageAccueil("/", nil, ""))

	if !strings.Contains(page, `id="local-kb-bar"`) {
		t.Errorf("barre des articles locaux absente de l'accueil :\n%s", page)
	}
	if !strings.Contains(page, "hidden max-w-5xl mx-auto px-6 pb-2") {
		t.Errorf("barre révélée par défaut (attendu : hidden) :\n%s", page)
	}
	if !strings.Contains(page, `id="kb-cards"`) {
		t.Errorf("grille partagée absente (injection des cartes locales impossible) :\n%s", page)
	}
	if !strings.Contains(page, `id="kb-empty"`) {
		t.Errorf("bascule de l'état vide absente :\n%s", page)
	}
	if !strings.Contains(page, "grz-local") {
		t.Errorf("script grz-local absent :\n%s", page)
	}
	if !strings.Contains(page, `id="kb-drawer"`) {
		t.Errorf("shell du tiroir absent (fiches locales attendues) :\n%s", page)
	}
}

// TestPageAccueilSharedGridKeepsCardsAndTable : avec des KB partagées, la
// grille porte les cartes serveur et le tableau son ancre d'injection
// (#kb-table-rows) ; la vue tableau reste cachée par défaut.
func TestPageAccueilSharedGridKeepsCardsAndTable(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB-1")}
	page := renderedPage(t, PageAccueil("/", kbs, ""))
	if !strings.Contains(page, `id="kb-cards"`) || strings.Contains(page, `data-tag-filtered="1"`) {
		t.Errorf("grille partagée invalide :\n%s", page)
	}
	if !strings.Contains(page, `id="kb-table-rows"`) {
		t.Errorf("ancre d'injection du tableau absente :\n%s", page)
	}
}

// TestViewToggleScriptComesAfterGrid : le script de bascule s'exécute au
// parsing : il ne doit apparaître qu'APRÈS la grille #kb-cards dans le DOM,
// sinon querySelector ne trouve pas la vue (grille null → abandon silencieux,
// les boutons cartes/tableau meurent — régression observée).
func TestViewToggleScriptComesAfterGrid(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB-1")}
	page := renderedPage(t, PageAccueil("/", kbs, ""))

	gridIdx := strings.Index(page, `id="kb-cards"`)
	scriptIdx := strings.Index(page, "grz-view")
	if gridIdx < 0 || scriptIdx < 0 {
		t.Fatalf("grille ou script de bascule absent (grid=%d script=%d)", gridIdx, scriptIdx)
	}
	if gridIdx > scriptIdx {
		t.Errorf("script de bascule exécuté avant la grille (grid@%d, script@%d) : boutons morts à l'exécution", gridIdx, scriptIdx)
	}
}

// TestLocalScriptComesAfterDrawerShell : le script grz-local s'exécute au
// parsing et ouvre les fiches locales dans le tiroir — le shell du tiroir
// doit apparaître AVANT lui dans le DOM. Régression observée : la barre
// (avec son script) était rendue avant le shell, drawer était capturé
// null, cliquer sur une carte locale ne faisait plus rien.
func TestLocalScriptComesAfterDrawerShell(t *testing.T) {
	page := renderedPage(t, PageAccueil("/", nil, ""))

	drawerIdx := strings.Index(page, `id="kb-drawer"`)
	scriptIdx := strings.Index(page, "grz-local")
	if drawerIdx < 0 || scriptIdx < 0 {
		t.Fatalf("shell du tiroir ou script local absent (drawer=%d script=%d)", drawerIdx, scriptIdx)
	}
	if drawerIdx > scriptIdx {
		t.Errorf("script grz-local exécuté avant le shell du tiroir (drawer@%d, script@%d) : fiches locales inouvrables", drawerIdx, scriptIdx)
	}
}

// TestPageAccueilTagFilterMarksGrid : un filtre tag actif marque la grille
// (data-tag-filtered) — le script suspend l'injection locale.
func TestPageAccueilTagFilterMarksGrid(t *testing.T) {
	kbs := []kb.KB{mustCardKB(t, "KB-1", "mm")}
	page := renderedPage(t, PageAccueil("/", kbs, "mm"))
	if !strings.Contains(page, `data-tag-filtered="1"`) {
		t.Errorf("marqueur de filtre absents de la grille :\n%s", page)
	}
}

// TestPageKBCreateCarriesLocalToggle : les deux formulaires de création
// portent la bascule locale (cachée par défaut), le marqueur
// data-local-save et le bouton de connexion ; l'édition d'une KB partagée
// reste purement serveur.
func TestPageKBCreateCarriesLocalToggle(t *testing.T) {
	for _, name := range []string{"création", "création libre"} {
		var page string
		if name == "création libre" {
			page = renderedPage(t, PageKBCreateQuickWith("/", FormulaireKB{}))
		} else {
			page = renderedPage(t, PageKBCreateWith("/", FormulaireKB{}))
		}
		if !strings.Contains(page, `id="kb-local-save"`) {
			t.Errorf("bascule locale absente de la %s :\n%s", name, page)
		}
		if !strings.Contains(page, `data-local-save="1"`) {
			t.Errorf("marqueur form[data-local-save] absent de la %s", name)
		}
		if !strings.Contains(page, "data-local-connect") {
			t.Errorf("bouton de connexion absent de la %s", name)
		}
	}
}

// TestPageKBEditHasNoLocalToggle : l'édition d'une KB partagée persiste
// côté serveur. Le périmètre local est création/édition de fichiers poste.
func TestPageKBEditHasNoLocalToggle(t *testing.T) {
	page := renderedPage(t, PageKBEdit("/", mustCardKB(t, "KB-1")))
	if strings.Contains(page, `id="kb-local-save"`) {
		t.Errorf("bascule locale indue sur l'édition partagée")
	}
}

// TestLocalKBScriptMergedGridContract : le script injecte dans la grille
// partagée et la vue tableau, purge par data-local-entry, suspend
// l'injection quand un filtre tag est actif, fais trembler l'état vide.
func TestLocalKBScriptMergedGridContract(t *testing.T) {
	for _, marker := range []string{
		"grz-local",
		"showDirectoryPicker",
		"isSecureContext",
		"grz-local-kb",
		"/kb/local/serialize",
		"/kb/local/parse",
		"/kb/local/render",
		"stopPropagation",
		"journal-",
		"data-local-entry",
		"kb-table-rows",
		"kb-empty",
		"data-tag-filtered",
		"grzFuzzyRefresh",
		// parité carte serveur : heading module, pied tags + T-code
		"afterFirstLine",
		// parité fiche serveur : clone du panel, Erreur trace brute mono
		"addLocalSection",
		"px-6 py-5 space-y-5",
		"whitespace-pre-wrap break-words",
		"/n",
		// édition locale : case cochée sans attendre le réseau
		"editingLocal",
		"toggle.checked = true",
	} {
		if !strings.Contains(localKBScript, marker) {
			t.Errorf("marqueur manquant dans le script grz-local : %q", marker)
		}
	}
}
