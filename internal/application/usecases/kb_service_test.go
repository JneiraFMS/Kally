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

func domainKB(t *testing.T, id string, tags ...string) domainkb.KB {
	t.Helper()
	item, err := domainkb.NewKB(id, "e", "n", "r", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return item
}
