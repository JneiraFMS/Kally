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

// TestKBServiceDetailReturnsBacklinksAndKnownIDs : la fiche expose les KB qui
// la citent avec [[...]] et l'ensemble des IDs connus, nécessaire au rendu
// des liens morts côté présentation.
func TestKBServiceDetailReturnsBacklinksAndKnownIDs(t *testing.T) {
	repo := newFakeRepository()
	a, err := domainkb.NewKB("KB-1", "", "voir [[KB-2]]", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	b, err := domainkb.NewKB("KB-2", "", "", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	c, err := domainkb.NewKB("KB-3", "", "[[KB-2]] et [[KB-9]]", "")
	if err != nil {
		t.Fatalf("NewKB(): %v", err)
	}
	repo.items["KB-1"] = a
	repo.items["KB-2"] = b
	repo.items["KB-3"] = c
	service := NewKBService(repo)

	detail, err := service.Detail(context.Background(), "KB-2")
	if err != nil {
		t.Fatalf("Detail() inattendu: %v", err)
	}
	if got, want := domainkb.IDs(detail.Backlinks), []string{"KB-1", "KB-3"}; !slices.Equal(got, want) {
		t.Errorf("Detail().Backlinks = %v, want %v", got, want)
	}
	if got, want := detail.KnownIDs, []string{"KB-1", "KB-2", "KB-3"}; !slices.Equal(got, want) {
		t.Errorf("Detail().KnownIDs = %v, want %v", got, want)
	}
	if got := detail.Neighbors; len(got) != 0 {
		t.Errorf("Detail(KB-2).Neighbors = %v, want vide (liens entrants seulement)", domainkb.IDs(got))
	}

	detail, err = service.Detail(context.Background(), "KB-3")
	if err != nil {
		t.Fatalf("Detail() inattendu: %v", err)
	}
	if got, want := domainkb.IDs(detail.Neighbors), []string{"KB-2"}; !slices.Equal(got, want) {
		t.Errorf("Detail(KB-3).Neighbors = %v, want %v (cible sortante existante)", got, want)
	}

	detail, err = service.Detail(context.Background(), "KB-1")
	if err != nil {
		t.Fatalf("Detail() inattendu: %v", err)
	}
	if len(detail.Backlinks) != 0 {
		t.Errorf("Detail(KB-1).Backlinks = %v, want vide", domainkb.IDs(detail.Backlinks))
	}
	if got, want := domainkb.IDs(detail.Neighbors), []string{"KB-2"}; !slices.Equal(got, want) {
		t.Errorf("Detail(KB-1).Neighbors = %v, want %v", got, want)
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

func domainKB(t *testing.T, id string, tags ...string) domainkb.KB {
	t.Helper()
	item, err := domainkb.NewKB(id, "e", "n", "r", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return item
}
