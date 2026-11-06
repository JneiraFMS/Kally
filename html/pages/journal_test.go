package pages

import (
	"strings"
	"testing"

	"grz/internal/application/dto"
	domainjournal "grz/internal/domain/journal"
)

// mustEntry construit une entrée valide pour les tests.
func mustEntry(t *testing.T, date, content string) domainjournal.Entry {
	t.Helper()
	entry, err := domainjournal.New(date, content)
	if err != nil {
		t.Fatalf("New(%q): %v", date, err)
	}
	return entry
}

// TestPageJournalNoteDuJour : la note du jour — bouton de création quand
// elle n'existe pas, lien vers sa fiche sinon.
func TestPageJournalNoteDuJour(t *testing.T) {
	html := renderedPage(t, PageJournal("/", dto.JournalPage{Today: "2026-11-03", TodayExists: false}))

	if !strings.Contains(html, `href="/journal/new"`) {
		t.Errorf("PageJournal() sans bouton de création:\n%s", html)
	}
	if !strings.Contains(html, ">Créer la note du jour<") {
		t.Errorf("PageJournal() sans libellé du bouton de création")
	}
	if strings.Contains(html, "Ouvrir la note du jour") {
		t.Errorf("PageJournal() propose d'ouvrir une note inexistante")
	}

	entryExists := true
	html = renderedPage(t, PageJournal("/", dto.JournalPage{Today: "2026-11-03", TodayExists: entryExists}))
	if !strings.Contains(html, "/journal/2026-11-03") {
		t.Errorf("PageJournal(existante) sans lien vers la note du jour:\n%s", html)
	}
	if !strings.Contains(html, ">Ouvrir la note du jour<") {
		t.Errorf("PageJournal(existante) sans libellé d'ouverture")
	}
	if strings.Contains(html, ">Créer la note du jour<") {
		t.Errorf("PageJournal(existante) propose encore de créer la note du jour")
	}
}

// TestPageJournalEntreesRecentesDAboard : les entrées du journal sont listées
// dans l'ordre fourni (plus récentes d'abord, préparé par le service), avec
// un extrait de leur markdown libre.
func TestPageJournalEntreesRecentesDavboard(t *testing.T) {
	entries := []domainjournal.Entry{
		mustEntry(t, "2026-11-03", "note recente"),
		mustEntry(t, "2026-01-15", "note ancienne"),
	}

	html := renderedPage(t, PageJournal("/", dto.JournalPage{Entries: entries}))

	// Contenu markdown : extrait en clair
	if !strings.Contains(html, "note recente") || !strings.Contains(html, "note ancienne") {
		t.Errorf("PageJournal() extrait absent:\n%s", html)
	}
	rec := strings.Index(html, "/journal/2026-11-03")
	anc := strings.Index(html, "/journal/2026-01-15")
	if rec < 0 || anc < 0 {
		t.Fatalf("PageJournal() entrée manquante:\n%s", html)
	}
	if rec > anc {
		t.Errorf("PageJournal() entrées dans le mauvais ordre (récentes d'abord)")
	}
	if !strings.Contains(html, "Entrées précédentes") {
		t.Errorf("PageJournal() sans titre de liste d'entrées")
	}
}

// TestPageJournalSearchBar : la liste d'entrées porte la barre de fuzzy
// partagée (attributs déclaratifs data-fuzzy-*) : champ, compteur, état
// « aucun résultat », index data-search par ligne et script client. Sans
// aucune entrée, ni barre ni état vide ne sont rendus.
func TestPageJournalSearchBar(t *testing.T) {
	entries := []domainjournal.Entry{mustEntry(t, "2026-11-03", "note recente")}
	html := renderedPage(t, PageJournal("/", dto.JournalPage{Entries: entries}))

	for _, want := range []string{
		`id="journal-search"`,
		`data-fuzzy`,
		`data-fuzzy-count="journal-search-count"`,
		`data-fuzzy-no-results="journal-no-results"`,
		`id="journal-no-results"`,
		// Clé d'index de la ligne : date brute, date en clair, extrait.
		`data-search="2026-11-03 3 novembre 2026 note recente"`,
		"grz-fuzzy",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("recherche du journal sans %q:\n%s", want, html)
		}
	}

	// Sans entrée : progressive enhancement inutile, ni barre ni état vide.
	empty := renderedPage(t, PageJournal("/", dto.JournalPage{}))
	if strings.Contains(empty, "journal-search") || strings.Contains(empty, "grz-fuzzy") {
		t.Errorf("barre de recherche rendue sans aucune entrée:\n%s", empty)
	}
}

// TestPageJournalAucuneEntree : état vide clair quand le journal est vide.
func TestPageJournalAucuneEntree(t *testing.T) {
	html := renderedPage(t, PageJournal("/", dto.JournalPage{}))

	// gomponents échappe l'apostrophe dans le texte sérialisé
	if !strings.Contains(html, "Pas encore d&#39;entrée") {
		t.Errorf("PageJournal() sans état vide:\n%s", html)
	}
	if strings.Contains(html, "Entrées précédentes") {
		t.Errorf("PageJournal(vide) affiche une liste d'entrées")
	}
}

// TestPageJournalNoNestedAnchors : aucun <a> imbriqué (HTML invalide).
func TestPageJournalNoNestedAnchors(t *testing.T) {
	page := dto.JournalPage{
		Today:       "2026-11-03",
		TodayExists: true,
		Entries:     []domainjournal.Entry{mustEntry(t, "2026-11-03", "e")},
	}
	noNestedAnchors(t, renderedPage(t, PageJournal("/", page)))
}

// TestJournalEntrySearchKey : la clé d'index d'une entrée réunit la date
// brute, la date en clair et les 3 premières lignes non vides du markdown
// (garde le DOM léger ; tout en minuscules, comme l'index des KB).
func TestJournalEntrySearchKey(t *testing.T) {
	entry := mustEntry(t, "2026-11-05",
		"Achats de consommables\n\nfournisseur Airbus Defense\nligne trois\nquatrième ligne")

	key := journalSearchKey(entry)
	for _, want := range []string{"2026-11-05", "5 novembre 2026", "achats de consommables", "fournisseur airbus defense", "ligne trois"} {
		if !strings.Contains(key, want) {
			t.Errorf("journalSearchKey() sans %q: %q", want, key)
		}
	}
	if strings.Contains(key, "quatrième") {
		t.Errorf("journalSearchKey() dépasse 3 lignes non vides: %q", key)
	}

	// Les lignes vide/espaces sont sautées ; une première ligne unique
	// donne la clé exacte « dates » + extrait.
	whitespace := mustEntry(t, "2026-01-01", "première ligne\n\n   \n\n   \n")
	if key := journalSearchKey(whitespace); key != "2026-01-01 1 janvier 2026 première ligne" {
		t.Errorf("journalSearchKey(lignes vides ignorées) = %q", key)
	}
}

// TestPageJournalCreate : le formulaire de création est libre — la date du
// jour est préselectionnée, le contenu est un textarea markdown, et les
// anciens champs KB (Identifiant, Erreur, Tags) sont absents.
func TestPageJournalCreate(t *testing.T) {
	html := renderedPage(t, PageJournalCreateWith("/", FormulaireJournal{
		Valeurs: map[string]string{"date": "2026-11-03"},
	}))

	if !strings.Contains(html, `type="date"`) || !strings.Contains(html, `value="2026-11-03"`) {
		t.Errorf("PageJournalCreate() sans champ date prérempli:\n%s", html)
	}
	if !strings.Contains(html, `name="content"`) || !strings.Contains(html, "<textarea") {
		t.Errorf("PageJournalCreate() sans textarea markdown")
	}
	if strings.Contains(html, "Identifiant de la KB") || strings.Contains(html, `name="id"`) {
		t.Errorf("PageJournalCreate() expose encore le formulaire KB (Identifiant/E Tags)")
	}
	if strings.Contains(html, `name="tags"`) {
		t.Errorf("PageJournalCreate() expose encore les tags")
	}
}

// TestPageJournalEdit : le formulaire d'édition fige la date et préremplit
// le contenu.
func TestPageJournalEdit(t *testing.T) {
	entry := mustEntry(t, "2026-11-03", "contenu note libre")
	html := renderedPage(t, PageJournalEdit("/", entry))

	if !strings.Contains(html, `value="2026-11-03"`) || !strings.Contains(html, "readonly") {
		t.Errorf("PageJournalEdit() sans date figée:\n%s", html)
	}
	if !strings.Contains(html, "contenu note libre") {
		t.Errorf("PageJournalEdit() sans contenu prérempli:\n%s", html)
	}
	if !strings.Contains(html, "3 novembre 2026") {
		t.Errorf("PageJournalEdit() sans date en clair dans le titre")
	}
}

// TestPageJournalEntry : la fiche affiche la date en clair, le markdown
// rendu, la navigation jour et les actions Modifier/Supprimer.
func TestPageJournalEntry(t *testing.T) {
	page := dto.JournalEntryPage{
		Entry: mustEntry(t, "2026-11-02", "une note avec du **gras**"),
		Prev:  "2026-11-01",
		Next:  "2026-11-03",
	}
	html := renderedPage(t, PageJournalEntry("/", page))

	if !strings.Contains(html, "2 novembre 2026") {
		t.Errorf("PageJournalEntry() sans date en clair:\n%s", html)
	}
	if !strings.Contains(html, "<strong>gras</strong>") {
		t.Errorf("PageJournalEntry() markdown libre non rendu:\n%s", html)
	}
	if !strings.Contains(html, `href="/journal/2026-11-01" class=`) {
		t.Errorf("PageJournalEntry() sans lien jour précédent")
	}
	if !strings.Contains(html, `href="/journal/2026-11-03" class=`) {
		t.Errorf("PageJournalEntry() sans lien jour suivant")
	}
	if !strings.Contains(html, "/journal/2026-11-02/edit") {
		t.Errorf("PageJournalEntry() sans action Modifier")
	}
	if !strings.Contains(html, `action="/journal/delete"`) || !strings.Contains(html, `name="date" value="2026-11-02"`) {
		t.Errorf("PageJournalEntry() sans suppression datée:\n%s", html)
	}
}

// TestPageJournalEntryNavigationExtrémités : un jour aux extrémités ne rend
// que le voisin existant (pourtour invisible, jamais de lien mort).
func TestPageJournalEntryNavigationExtremites(t *testing.T) {
	page := dto.JournalEntryPage{
		Entry: mustEntry(t, "2026-10-30", "la plus ancienne"),
		Prev:  "",
		Next:  "2026-11-01",
	}
	html := renderedPage(t, PageJournalEntry("/", page))

	links := strings.Count(html, "/journal/2026-10-30")
	if links < 1 {
		t.Errorf("PageJournalEntry() le lien vers la note courante manque sur la suppression")
	}
	// Pas de lien mort vers un voisin absent
	if strings.Contains(html, `href="/journal/""`) {
		t.Errorf("PageJournalEntry() lien vers une date vide")
	}
	// Le voisin suivant existe et est cliquable
	if !strings.Contains(html, `href="/journal/2026-11-01" class=`) {
		t.Errorf("PageJournalEntry() sans lien jour suivant")
	}
}

// TestPageJournalNoNestedAnchorsEntry : aucun <a> imbriqué dans la fiche.
func TestPageJournalEntryNoNestedAnchors(t *testing.T) {
	page := dto.JournalEntryPage{
		Entry: mustEntry(t, "2026-11-02", "**note**"),
		Prev:  "2026-11-01",
		Next:  "2026-11-03",
	}
	noNestedAnchors(t, renderedPage(t, PageJournalEntry("/", page)))
}

// TestFormatDateFR : dates canoniques en clair, non canoniques rendues
// brute — jamais d'index hors bornes.
func TestFormatDateFR(t *testing.T) {
	if got := formatDateFR("2026-11-03"); got != "3 novembre 2026" {
		t.Errorf("formatDateFR() = %q", got)
	}
	if got := formatDateFR("2026-01-01"); got != "1 janvier 2026" {
		t.Errorf("formatDateFR() = %q, want sans zéro initial", got)
	}
	for _, raw := range []string{"", "brouillon", "2026-00-01", "2026-13-01", "2026-01-32"} {
		if got := formatDateFR(raw); got != raw {
			t.Errorf("formatDateFR(%q) = %q, want brut", raw, got)
		}
	}
}
