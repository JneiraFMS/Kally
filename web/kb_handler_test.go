package web

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"grz/internal/kb"
)

// fakeRepo est un double simple du port kb.Repository pour les tests.
type fakeRepo struct {
	saved map[string]kb.KB
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{saved: map[string]kb.KB{}}
}

func (f *fakeRepo) Save(_ context.Context, k kb.KB) error {
	f.saved[k.ID] = k
	return nil
}

func (f *fakeRepo) FindAll(_ context.Context) ([]kb.KB, error) {
	kbs := []kb.KB{}
	for _, kb := range f.saved {
		kbs = append(kbs, kb)
	}
	return kbs, nil
}

func (f *fakeRepo) FindByID(_ context.Context, id string) (*kb.KB, error) {
	if b, ok := f.saved[id]; ok {
		return &b, nil
	}
	return nil, fmt.Errorf("kb: %w: %s", kb.ErrNotFound, id)
}

func TestKBHandlerCreateSavesKB(t *testing.T) {
	repo := newFakeRepo()
	h := NewKBHandler(repo)

	form := url.Values{
		"id":         {"KB-0001"},
		"erreur":     {"Div par zéro"},
		"notes":      {"Tester le diviseur"},
		"resolution": {"Vérifier avant division"},
	}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != 200 {
		t.Errorf("Create() status = %d, want 200", w.Code)
	}
	if _, ok := repo.saved["KB-0001"]; !ok {
		t.Errorf("Create() n'a pas sauvegardé la KB, saved = %v", repo.saved)
	}
}

func TestKBHandlerCreateRejectsInvalidID(t *testing.T) {
	repo := newFakeRepo()
	h := NewKBHandler(repo)

	form := url.Values{"id": {"../hack"}, "erreur": {"e"}, "notes": {"n"}, "resolution": {"r"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != 400 {
		t.Errorf("Create() status = %d, want 400", w.Code)
	}
	if len(repo.saved) != 0 {
		t.Errorf("Create() a sauvegardé une KB invalide: %v", repo.saved)
	}
}

func seedRepo(t *testing.T, ids ...string) *fakeRepo {
	t.Helper()
	repo := newFakeRepo()
	for _, id := range ids {
		k, err := kb.NewKB(id, "erreur démo "+id, "notes", "résolution")
		if err != nil {
			t.Fatalf("NewKB(%q): %v", id, err)
		}
		if err := repo.Save(context.Background(), k); err != nil {
			t.Fatalf("Save(%q): %v", id, err)
		}
	}
	return repo
}

func TestKBHandlerHomeShowsKBCards(t *testing.T) {
	repo := seedRepo(t, "KB-0001", "KB-0002")
	h := NewKBHandler(repo)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)

	if w.Code != 200 {
		t.Errorf("Home() status = %d, want 200", w.Code)
	}
	for _, id := range []string{"KB-0001", "KB-0002"} {
		if !strings.Contains(w.Body.String(), id) {
			t.Errorf("Home() n'affiche pas la carte de %s", id)
		}
	}
	if strings.Contains(w.Body.String(), "<form") {
		t.Errorf("Home() ne doit plus contenir de formulaire")
	}
}

func TestKBHandlerHomeEmptyState(t *testing.T) {
	h := NewKBHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)

	if !strings.Contains(w.Body.String(), "Aucune KB") {
		t.Errorf("Home() ne gère pas l'état vide")
	}
}
