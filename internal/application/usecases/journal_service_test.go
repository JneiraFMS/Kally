package usecases

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"grz/internal/application"
	"grz/internal/application/dto"
	domainjournal "grz/internal/domain/journal"
)

// fakeJournalRepository est un double simple du port de persistance du
// journal pour les tests du use case.
type fakeJournalRepository struct {
	saved map[string]domainjournal.Entry
}

func newFakeJournalRepository() *fakeJournalRepository {
	return &fakeJournalRepository{saved: map[string]domainjournal.Entry{}}
}

func (f *fakeJournalRepository) Save(_ context.Context, entry domainjournal.Entry) error {
	f.saved[entry.Date] = entry
	return nil
}

func (f *fakeJournalRepository) Exists(_ context.Context, date string) (bool, error) {
	_, ok := f.saved[date]
	return ok, nil
}

func (f *fakeJournalRepository) ByDate(_ context.Context, date string) (*domainjournal.Entry, error) {
	if entry, ok := f.saved[date]; ok {
		return &entry, nil
	}
	return nil, fmt.Errorf("%w: %s", domainjournal.ErrNotFound, date)
}

func (f *fakeJournalRepository) Entries(_ context.Context) ([]domainjournal.Entry, error) {
	entries := make([]domainjournal.Entry, 0, len(f.saved))
	for _, entry := range f.saved {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (f *fakeJournalRepository) Delete(_ context.Context, date string) error {
	if _, ok := f.saved[date]; !ok {
		return fmt.Errorf("%w: %s", domainjournal.ErrNotFound, date)
	}
	delete(f.saved, date)
	return nil
}

// TestJournalServicePage : le journal liste les entrées de la plus récente
// à la plus ancienne et signale si la note du jour existe (la date du jour
// vient de l'adapter → use case déterministe).
func TestJournalServicePage(t *testing.T) {
	repo := newFakeJournalRepository()
	svc := NewJournalService(repo)
	ctx := context.Background()

	page, err := svc.Page(ctx, "2026-11-03")
	if err != nil {
		t.Fatalf("Page() inattendu: %v", err)
	}
	if page.Today != "2026-11-03" || page.TodayExists {
		t.Errorf("Page() journal vide = %+v, want Today=2026-11-03 absent", page)
	}

	// Deux créations, la page les liste recentes d'abord, sans la KB du jour
	must := func(date, content string) dto.JournalInput {
		if _, err := svc.Create(ctx, dto.JournalInput{Date: date, Content: content}); err != nil {
			t.Fatalf("Create(%q): %v", date, err)
		}
		return dto.JournalInput{}
	}
	must("2026-11-03", "note du jour")
	must("2026-01-15", "note ancienne")

	page, err = svc.Page(ctx, "2026-11-03")
	if err != nil {
		t.Fatalf("Page() inattendu: %v", err)
	}
	if !page.TodayExists {
		t.Errorf("Page() ne signale pas la note du jour existante")
	}
	if len(page.Entries) != 2 || page.Entries[0].Date != "2026-11-03" || page.Entries[1].Date != "2026-01-15" {
		t.Errorf("Page() entrées = %+v, want 2026-11-03 puis 2026-01-15", page.Entries)
	}
}

// TestJournalServiceCreateAndDuplicate : créer pour un jour couvert
// échoue (ErrAlreadyExists) sans écraser la note existante.
func TestJournalServiceCreateAndDuplicate(t *testing.T) {
	repo := newFakeJournalRepository()
	svc := NewJournalService(repo)
	ctx := context.Background()

	created, err := svc.Create(ctx, dto.JournalInput{Date: "2026-11-03", Content: "**note** du jour"})
	if err != nil {
		t.Fatalf("Create() inattendu: %v", err)
	}
	if created.Content != "**note** du jour" {
		t.Errorf("Create() content = %q, want conservé tel quel", created.Content)
	}

	_, err = svc.Create(ctx, dto.JournalInput{Date: "2026-11-03", Content: "autre"})
	if !errors.Is(err, application.ErrAlreadyExists) {
		t.Errorf("Create() doublon = %v, want ErrAlreadyExists", err)
	}
	if got := repo.saved["2026-11-03"].Content; got != "**note** du jour" {
		t.Errorf("Create() doublon a écrasé le contenu: %q", got)
	}
}

// TestJournalServiceCreateValide : date invalide ou contenu vide refusés.
func TestJournalServiceCreateValide(t *testing.T) {
	svc := NewJournalService(newFakeJournalRepository())
	ctx := context.Background()

	if _, err := svc.Create(ctx, dto.JournalInput{Date: "2026-02-30", Content: "x"}); !errors.Is(err, domainjournal.ErrInvalidDate) {
		t.Errorf("Create() 30 février = %v, want ErrInvalidDate", err)
	}
	if _, err := svc.Create(ctx, dto.JournalInput{Date: "2026-11-03", Content: "   "}); !errors.Is(err, domainjournal.ErrEmptyContent) {
		t.Errorf("Create() contenu blanc = %v, want ErrEmptyContent", err)
	}
}

// TestJournalServiceEntryAndNavigation : la fiche d'une entrée expose la
// date des jours précédent (plus ancien) et suivant (plus récent).
func TestJournalServiceEntryAndNavigation(t *testing.T) {
	repo := newFakeJournalRepository()
	svc := NewJournalService(repo)
	ctx := context.Background()

	for _, spec := range []struct{ date, content string }{
		{"2026-11-03", "recente"}, {"2026-11-02", "milieu"}, {"2026-10-30", "ancienne"},
	} {
		if _, err := svc.Create(ctx, dto.JournalInput{Date: spec.date, Content: spec.content}); err != nil {
			t.Fatalf("Create(%q): %v", spec.date, err)
		}
	}

	page, err := svc.Entry(ctx, "2026-11-02")
	if err != nil {
		t.Fatalf("Entry() inattendu: %v", err)
	}
	if page.Entry.Content != "milieu" || page.Prev != "2026-10-30" || page.Next != "2026-11-03" {
		t.Errorf("Entry() = %+v, want content milieu prev 2026-10-30 next 2026-11-03", page)
	}

	// Extrémités
	page, err = svc.Entry(ctx, "2026-10-30")
	if err != nil {
		t.Fatalf("Entry() inattendu: %v", err)
	}
	if page.Prev != "" || page.Next != "2026-11-02" {
		t.Errorf("Entry() extrémité ancienne = (%q, %q), want («» , 2026-11-02)", page.Prev, page.Next)
	}

	// Absente
	if _, err := svc.Entry(ctx, "2026-09-01"); !errors.Is(err, domainjournal.ErrNotFound) {
		t.Errorf("Entry() absent = %v, want ErrNotFound", err)
	}
}

// TestJournalServiceUpdateAndDelete : mettre à jour une entrée ré-enregistre
// son contenu (Save upserte par date) ; supprimer une entrée absente
// échoue, une entrée présente est retirée.
func TestJournalServiceUpdateAndDelete(t *testing.T) {
	repo := newFakeJournalRepository()
	svc := NewJournalService(repo)
	ctx := context.Background()

	if _, err := svc.Create(ctx, dto.JournalInput{Date: "2026-11-03", Content: "v1"}); err != nil {
		t.Fatalf("Create(): %v", err)
	}

	if _, err := svc.Update(ctx, dto.JournalInput{Date: "2026-11-03", Content: "v2"}); err != nil {
		t.Fatalf("Update() inattendu: %v", err)
	}
	if got := repo.saved["2026-11-03"].Content; got != "v2" {
		t.Errorf("Update() contenu = %q, want v2", got)
	}
	if _, err := svc.Update(ctx, dto.JournalInput{Date: "2026-11-03", Content: "   "}); !errors.Is(err, domainjournal.ErrEmptyContent) {
		t.Errorf("Update() contenu blanc = %v, want ErrEmptyContent", err)
	}

	if err := svc.Delete(ctx, "2026-11-03"); err != nil {
		t.Fatalf("Delete() inattendu: %v", err)
	}
	if _, ok := repo.saved["2026-11-03"]; ok {
		t.Errorf("Delete() a laissé l'entrée: %v", repo.saved)
	}
	if err := svc.Delete(ctx, "2026-11-03"); !errors.Is(err, domainjournal.ErrNotFound) {
		t.Errorf("Delete() absent = %v, want ErrNotFound", err)
	}
}
