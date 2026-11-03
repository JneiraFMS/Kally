package kb

import (
	"slices"
	"testing"
)

func mustJournalKB(t *testing.T, id string, tags ...string) KB {
	t.Helper()
	k, err := NewKB(id, "e", "n", "r", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return k
}

// TestFilterByIDPrefix : seules les KB dont l'ID commence par le préfixe
// sont conservées, triées de la plus récente (ID le plus grand) à la plus
// ancienne — l'ordre d'édition du journal Logseq.
func TestFilterByIDPrefix(t *testing.T) {
	items := []KB{
		mustJournalKB(t, "journal-2026-01-01", "journal"),
		mustJournalKB(t, "KB-0009", "calcul"),
		mustJournalKB(t, "journal-2026-11-03", "journal"),
		mustJournalKB(t, "journal-2026-02-15", "journal"),
	}

	got := make([]string, 0, len(items))
	for _, k := range FilterByIDPrefix(items, "journal-") {
		got = append(got, k.ID)
	}
	want := []string{"journal-2026-11-03", "journal-2026-02-15", "journal-2026-01-01"}
	if !slices.Equal(got, want) {
		t.Errorf("FilterByIDPrefix() = %v, want %v", got, want)
	}
}

// TestFilterByIDPrefixVide : préfixe qui ne matche rien, préfixe vide, ou
// aucune entrée — résultat vide, jamais nil panique.
func TestFilterByIDPrefixVide(t *testing.T) {
	items := []KB{mustJournalKB(t, "KB-0001")}

	if got := IDs(FilterByIDPrefix(items, "journal-")); len(got) != 0 {
		t.Errorf("FilterByIDPrefix() préfixe absent = %v, want vide", got)
	}
	if got := IDs(FilterByIDPrefix(items, "")); len(got) != 0 {
		t.Errorf("FilterByIDPrefix() préfixe vide = %v, want vide", got)
	}
	if got := IDs(FilterByIDPrefix(nil, "journal-")); len(got) != 0 {
		t.Errorf("FilterByIDPrefix() aucune KB = %v, want vide", got)
	}
}

// TestNormalizeTagFilter : le filtre de tags issu d'une URL est nettoyé —
// minuscules, espaces coupés, doublons et vides retirés. Un tag aux
// caractères interdits est simplement ignoré (pas d'erreur : le filtre
// utilisateur ne doit jamais renvoyer 400, juste ne rien filtrer).
func TestNormalizeTagFilter(t *testing.T) {
	got := NormalizeTagFilter([]string{"  Incident ", "KPI", "", "incident", "a%b"})
	want := []string{"incident", "kpi"}
	if !slices.Equal(got, want) {
		t.Errorf("NormalizeTagFilter() = %v, want %v", got, want)
	}

	if got := NormalizeTagFilter(nil); len(got) != 0 {
		t.Errorf("NormalizeTagFilter(nil) = %v, want vide", got)
	}
}
