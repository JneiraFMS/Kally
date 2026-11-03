package journal

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"
)

func mustEntry(t *testing.T, date, content string) Entry {
	t.Helper()
	entry, err := New(date, content)
	if err != nil {
		t.Fatalf("New(%q): %v", date, err)
	}
	return entry
}

// TestNew : une entrée de journal = une date canonique + du markdown libre.
// Pas d'ID (la date est l'identité), pas de tags.
func TestNew(t *testing.T) {
	entry, err := New("2026-11-03", "**Une note** libre\n\n- point un\n- point deux")
	if err != nil {
		t.Fatalf("New() inattendu: %v", err)
	}
	if entry.Date != "2026-11-03" || entry.Content != "**Une note** libre\n\n- point un\n- point deux" {
		t.Errorf("New() = %+v, want date 2026-11-03 + contenu tel quel", entry)
	}
}

// TestNewDateInvalide : tout ce qui ne respecte pas YYYY-MM-DD est refusé
// (séparateurs, hors bornes — 30 février —, inversion jour/mois ou
// chemin de directory traversal).
func TestNewDateInvalide(t *testing.T) {
	for _, date := range []string{"", "2026/11/03", "2026-02-30", "03-11-2026", "2026-01-01/..", "journal", "2026-1-1"} {
		if _, err := New(date, "contenu"); !errors.Is(err, ErrInvalidDate) {
			t.Errorf("New(%q) error = %v, want ErrInvalidDate", date, err)
		}
	}
	if err := ValidateDate("2026-02-30"); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("ValidateDate(30 février) = %v, want ErrInvalidDate", err)
	}
}

// TestNewContentVide : un contenu vide (ou blanc) est refusé — une note de
// journal sans texte n'a aucune raison d'être persistée.
func TestNewContentVide(t *testing.T) {
	if _, err := New("2026-11-03", ""); !errors.Is(err, ErrEmptyContent) {
		t.Errorf("New() content vide = %v, want ErrEmptyContent", err)
	}
	if _, err := New("2026-11-03", "   \n  "); !errors.Is(err, ErrEmptyContent) {
		t.Errorf("New() content blanc = %v, want ErrEmptyContent", err)
	}
}

// TestValidateDateInvalideFmt : l'erreur embarque la valeur fautive.
func TestValidateDateInvalideFmt(t *testing.T) {
	err := ValidateDate("brouillon")
	if err == nil || !errors.Is(err, ErrInvalidDate) || !strings.Contains(err.Error(), "brouillon") {
		t.Errorf("ValidateDate() = %v, want ErrInvalidDate mentionnant la valeur", err)
	}
}

// TestEntriesDesc : les entrées sont triées de la plus récente à la plus
// ancienne, sans muter la source (les dates ISO se comparent lexicalement).
func TestEntriesDesc(t *testing.T) {
	entries := []Entry{
		mustEntry(t, "2026-01-15", "ancienne"),
		mustEntry(t, "2026-11-03", "recente"),
		mustEntry(t, "2026-02-10", "milieu"),
	}
	source := slices.Clone(entries)

	got := EntriesDesc(entries)
	want := []string{"2026-11-03", "2026-02-10", "2026-01-15"}
	for i, entry := range got {
		if entry.Date != want[i] {
			t.Errorf("EntriesDesc() position %d = %q, want %q", i, entry.Date, want[i])
		}
	}
	if !slices.Equal(entries, source) {
		t.Errorf("EntriesDesc() a muté la liste source")
	}
}

// TestNeighborDates : navigation journal du jour précédent (plus ancien) et
// du jour suivant (plus récent), comme un journal Logseq.
func TestNeighborDates(t *testing.T) {
	entries := []Entry{
		mustEntry(t, "2026-11-03", "e"),
		mustEntry(t, "2026-11-02", "e"),
		mustEntry(t, "2026-11-01", "e"),
		mustEntry(t, "2026-10-30", "e"),
	}

	cases := []struct {
		date       string
		prev, next string
	}{
		{"2026-11-01", "2026-10-30", "2026-11-02"}, // milieu : les deux
		{"2026-11-03", "2026-11-02", ""},           // la plus récente : pas de suivant
		{"2026-10-30", "", "2026-11-01"},           // la plus ancienne : pas de précédent
		{"2026-09-01", "", ""},                     // absente
	}
	for _, tt := range cases {
		prev, next := NeighborDates(entries, tt.date)
		if prev != tt.prev || next != tt.next {
			t.Errorf("NeighborDates(%q) = (%q, %q), want (%q, %q)", tt.date, prev, next, tt.prev, tt.next)
		}
	}
}

// TestDateLayoutConst : la disposition de date canonique est exposée et
// vaut exactement ce que le domaine accepte.
func TestDateLayoutConst(t *testing.T) {
	if _, err := time.Parse(DateLayout, "2026-11-03"); err != nil {
		t.Errorf("DateLayout=%q n'accepte pas une date canonique: %v", DateLayout, err)
	}
}
