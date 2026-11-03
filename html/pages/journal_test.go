package pages

import (
	"net/url"
	"strings"
	"testing"

	"grz/internal/application/dto"
	kb "grz/internal/domain/kb"
)

func mustEntryKB(t *testing.T, id string, erreur string, tags ...string) kb.KB {
	t.Helper()
	k, err := kb.NewKB(id, erreur, "notes", "résolution", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return k
}

// TestPageJournalInstructionDuJour : la note du jour — bouton de création
// préremplie (id journal-<date> + tag journal) quand elle n'existe pas,
// lien vers la fiche sinon.
func TestPageJournalNoteDuJour(t *testing.T) {
	jour := "2026-11-03"
	todayID := "journal-" + jour

	html := renderedPage(t, PageJournal("/", dto.JournalPage{Today: jour, TodayID: todayID}))
	cible := "/kb/new?id=" + url.QueryEscape(todayID) + "&amp;tags=journal"
	if !strings.Contains(html, cible) {
		t.Errorf("PageJournal() sans bouton de création préremplie (%q):\n%s", cible, html)
	}
	if strings.Contains(html, "/journal-2026") && !strings.Contains(html, todayID) {
		t.Errorf("PageJournal() affiche une autre entrée que la note du jour")
	}

	entry := mustEntryKB(t, todayID, "panneau de substitution", "journal")
	html = renderedPage(t, PageJournal("/", dto.JournalPage{
		Today:       jour,
		TodayID:     todayID,
		TodayExists: true,
		TodayEntry:  &entry,
	}))
	if !strings.Contains(html, "/kb/detail?id="+todayID) {
		t.Errorf("PageJournal(existante) sans lien vers la fiche de la note du jour:\n%s", html)
	}
	if !strings.Contains(html, "panneau de substitution") {
		t.Errorf("PageJournal(existante) sans extrait de la note du jour")
	}
}

// TestPageJournalEntreesRecentesDAboard : les entrées du journal sont listées
// dans l'ordre fourni (plus récentes d'abord, préparé par le service),
// seules — les autres KB n'apparaissent jamais.
func TestPageJournalEntreesRecentesDAboard(t *testing.T) {
	entries := []kb.KB{
		mustEntryKB(t, "journal-2026-11-03", "note recente", "journal"),
		mustEntryKB(t, "journal-2026-01-15", "note ancienne", "journal"),
	}

	html := renderedPage(t, PageJournal("/", dto.JournalPage{Entries: entries}))

	rec, anc := strings.Index(html, "/kb/detail?id=journal-2026-11-03"), strings.Index(html, "/kb/detail?id=journal-2026-01-15")
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

// TestPageJournalAucuneEntree : état vide clair quand le journal est vide.
func TestPageJournalAucuneEntree(t *testing.T) {
	html := renderedPage(t, PageJournal("/", dto.JournalPage{}))

	// gomponents échappe l'apostrophe dans le texte sérialisé
	if !strings.Contains(html, "Pas encore d&#39;entrée") {
		t.Errorf("PageJournal() sans état vide:\n%s", html)
	}
	if strings.Contains(html, "/kb/detail?id=journal-") {
		t.Errorf("PageJournal(vide) liste des entrées inexistantes")
	}
}

// TestPageJournalNoNestedAnchors : aucun <a> imbriqué (HTML invalide).
func TestPageJournalNoNestedAnchors(t *testing.T) {
	entry := mustEntryKB(t, "journal-2026-11-03", "e", "journal")
	page := dto.JournalPage{
		Today:       "2026-11-03",
		TodayID:     "journal-2026-11-03",
		TodayExists: true,
		TodayEntry:  &entry,
		Entries:     []kb.KB{entry},
	}
	noNestedAnchors(t, renderedPage(t, PageJournal("/", page)))
}
