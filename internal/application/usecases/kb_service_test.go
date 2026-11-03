package usecases

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"grz/internal/application"
	"grz/internal/application/dto"
	domainkb "grz/internal/domain/kb"
)

type fakeRepository struct {
	items map[string]domainkb.KB
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{items: make(map[string]domainkb.KB)}
}

func (r *fakeRepository) Save(_ context.Context, item domainkb.KB) error {
	r.items[item.ID] = item
	return nil
}

func (r *fakeRepository) Exists(_ context.Context, id string) (bool, error) {
	_, ok := r.items[id]
	return ok, nil
}

func (r *fakeRepository) FindAll(_ context.Context) ([]domainkb.KB, error) {
	items := make([]domainkb.KB, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, item)
	}
	return items, nil
}

func (r *fakeRepository) FindByID(_ context.Context, id string) (*domainkb.KB, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domainkb.ErrNotFound, id)
	}
	return &item, nil
}

func (r *fakeRepository) Delete(_ context.Context, id string) error {
	if _, ok := r.items[id]; !ok {
		return fmt.Errorf("%w: %s", domainkb.ErrNotFound, id)
	}
	delete(r.items, id)
	return nil
}

func TestKBServiceCreateValidatesAndRejectsDuplicates(t *testing.T) {
	repo := newFakeRepository()
	service := NewKBService(repo)
	input := dto.KBInput{ID: "KB-1", Erreur: "e", Notes: "n", Resolution: "r", Tags: "Incident, kpi"}

	created, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("Create() inattendu: %v", err)
	}
	if !slices.Equal(created.Tags, []string{"incident", "kpi"}) {
		t.Errorf("Create() tags = %v, want [incident kpi]", created.Tags)
	}

	if _, err := service.Create(context.Background(), input); !errors.Is(err, application.ErrAlreadyExists) {
		t.Errorf("Create() doublon error = %v, want ErrAlreadyExists", err)
	}
	if _, err := service.Create(context.Background(), dto.KBInput{ID: "../bad"}); !errors.Is(err, domainkb.ErrInvalidID) {
		t.Errorf("Create() ID invalide error = %v, want ErrInvalidID", err)
	}
}

func TestKBServiceListFiltersByTag(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-1"] = domainKB(t, "KB-1", "incident")
	repo.items["KB-2"] = domainKB(t, "KB-2", "kpi")
	service := NewKBService(repo)

	got, err := service.List(context.Background(), "incident")
	if err != nil {
		t.Fatalf("List() inattendu: %v", err)
	}
	if ids := domainkb.IDs(got); !slices.Equal(ids, []string{"KB-1"}) {
		t.Errorf("List() IDs = %v, want [KB-1]", ids)
	}
}

func TestKBServiceDetailReturnsRelatedKBs(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-1"] = domainKB(t, "KB-1", "incident")
	repo.items["KB-2"] = domainKB(t, "KB-2", "incident")
	repo.items["KB-3"] = domainKB(t, "KB-3", "kpi")
	service := NewKBService(repo)

	detail, err := service.Detail(context.Background(), "KB-1")
	if err != nil {
		t.Fatalf("Detail() inattendu: %v", err)
	}
	if detail.KB.ID != "KB-1" || !slices.Equal(domainkb.IDs(detail.Related), []string{"KB-2"}) {
		t.Errorf("Detail() = %+v, want KB-1 related to KB-2", detail)
	}

	if _, err := service.Detail(context.Background(), "missing"); !errors.Is(err, domainkb.ErrNotFound) {
		t.Errorf("Detail() absent error = %v, want ErrNotFound", err)
	}
}

func TestKBServiceUpdateAndDelete(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-1"] = domainKB(t, "KB-1", "old")
	service := NewKBService(repo)

	updated, err := service.Update(context.Background(), dto.KBInput{ID: "KB-1", Erreur: "updated"})
	if err != nil {
		t.Fatalf("Update() inattendu: %v", err)
	}
	if updated.Erreur != "updated" {
		t.Errorf("Update() Erreur = %q, want updated", updated.Erreur)
	}
	if err := service.Delete(context.Background(), "KB-1"); err != nil {
		t.Fatalf("Delete() inattendu: %v", err)
	}
	if len(repo.items) != 0 {
		t.Errorf("Delete() a laissé des éléments: %v", repo.items)
	}
}

// TestKBServiceListFiltreAnd : plusieurs tags = requête AND (à la Logseq)
// — une KB doit porter TOUS les tags. Les tags brouillons (casse, espaces,
// doublons) sont normalisés avant le filtrage : le filtre ne rejette jamais.
func TestKBServiceListFiltreAnd(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-1"] = domainKB(t, "KB-1", "incident", "kpi")
	repo.items["KB-2"] = domainKB(t, "KB-2", "incident")
	repo.items["KB-3"] = domainKB(t, "KB-3", "kpi")
	service := NewKBService(repo)

	got, err := service.List(context.Background(), "incident", "kpi")
	if err != nil {
		t.Fatalf("List() inattendu: %v", err)
	}
	if ids := domainkb.IDs(got); !slices.Equal(ids, []string{"KB-1"}) {
		t.Errorf("List(incident, kpi) IDs = %v, want [KB-1]", ids)
	}

	// Normalisation : casse et doublons ne changent pas le résultat
	got, err = service.List(context.Background(), " Incident ", "KPI", "incident")
	if err != nil {
		t.Fatalf("List() normalisée inattendu: %v", err)
	}
	if ids := domainkb.IDs(got); !slices.Equal(ids, []string{"KB-1"}) {
		t.Errorf("List() normalisée IDs = %v, want [KB-1]", ids)
	}
}

// TestKBServiceJournal : le journal liste les KB préfixées journal-, de la
// plus récente à la plus ancienne, et expose la note du jour quand elle
// existe (sinon TodayEntry est nil et TodayExists est faux).
func TestKBServiceJournal(t *testing.T) {
	repo := newFakeRepository()
	repo.items["KB-9"] = domainKB(t, "KB-9", "calcul")
	repo.items["journal-2026-01-01"] = domainKB(t, "journal-2026-01-01", "journal")
	repo.items["journal-2026-11-03"] = domainKB(t, "journal-2026-11-03", "journal")
	service := NewKBService(repo)

	ctx := context.Background()
	page, err := service.Journal(ctx, "2026-11-02")
	if err != nil {
		t.Fatalf("Journal() inattendu: %v", err)
	}

	if page.Today != "2026-11-02" || page.TodayID != "journal-2026-11-02" {
		t.Errorf("Journal() Today = %q, TodayID = %q, want 2026-11-02 / journal-2026-11-02", page.Today, page.TodayID)
	}
	if page.TodayExists || page.TodayEntry != nil {
		t.Errorf("Journal() la note du jour (2026-11-02) n'existe pas : TodayExists=%v TodayEntry=%+v", page.TodayExists, page.TodayEntry)
	}
	want := []string{"journal-2026-11-03", "journal-2026-01-01"}
	if !slices.Equal(domainkbIDs(page.Entries), want) {
		t.Errorf("Journal() entrées = %v, want %v (ordres décroissant)", domainkbIDs(page.Entries), want)
	}

	// La note du jour existe : exposée via TodayEntry
	page, err = service.Journal(ctx, "2026-11-03")
	if err != nil {
		t.Fatalf("Journal() inattendu: %v", err)
	}
	if !page.TodayExists || page.TodayEntry == nil || page.TodayEntry.ID != "journal-2026-11-03" {
		t.Errorf("Journal() note du jour = %+v, want journal-2026-11-03", page.TodayEntry)
	}
}

// domainkbIDs extrait les IDs d'entrées sans trier (le helper IDs du
// domaine trierait et masquerait l'ordre du journal).
func domainkbIDs(items []domainkb.KB) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func domainKB(t *testing.T, id string, tags ...string) domainkb.KB {
	t.Helper()
	item, err := domainkb.NewKB(id, "e", "n", "r", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return item
}
