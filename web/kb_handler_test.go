package web

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"slices"
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
		if err := seedKB(t, repo, id); err != nil {
			t.Fatalf("seedKB(%q): %v", id, err)
		}
	}
	return repo
}

func seedKB(t *testing.T, repo *fakeRepo, id string, tags ...string) error {
	t.Helper()
	k, err := kb.NewKB(id, "erreur démo "+id, "notes", "résolution", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return repo.Save(context.Background(), k)
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

func TestKBHandlerCreateSavesTags(t *testing.T) {
	repo := newFakeRepo()
	h := NewKBHandler(repo)

	form := url.Values{
		"id":         {"KB-0001"},
		"erreur":     {"e"},
		"notes":      {"n"},
		"resolution": {"r"},
		"tags":       {" Incident, KPI "},
	}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != 200 {
		t.Errorf("Create() status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	saved := repo.saved["KB-0001"]
	if !slices.Equal(saved.Tags, []string{"incident", "kpi"}) {
		t.Errorf("Create() Tags = %v, want [incident kpi]", saved.Tags)
	}
}

func TestKBHandlerListFiltersByTag(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0002", "csv"); err != nil {
		t.Fatal(err)
	}
	h := NewKBHandler(repo)

	req := httptest.NewRequest("GET", "/kb?tag=kpi", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "KB-0001") {
		t.Errorf("List(filter kpi) n'affiche pas KB-0001")
	}
	if strings.Contains(body, "KB-0002") {
		t.Errorf("List(filter kpi) affiche KB-0002 hors filtre")
	}

	// Sans filtre : toutes les KB
	req = httptest.NewRequest("GET", "/kb", nil)
	w = httptest.NewRecorder()
	h.List(w, req)
	if !strings.Contains(w.Body.String(), "KB-0002") {
		t.Errorf("List(sans filtre) doit afficher toutes les KB")
	}
}

func TestKBHandlerShowGraph(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0002", "kpi"); err != nil {
		t.Fatal(err)
	}
	h := NewKBHandler(repo)

	req := httptest.NewRequest("GET", "/kb/graph", nil)
	w := httptest.NewRecorder()
	h.ShowGraph(w, req)

	if w.Code != 200 {
		t.Errorf("ShowGraph() status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{"<svg", "KB-0001", "KB-0002", "<line"} {
		if !strings.Contains(body, want) {
			t.Errorf("ShowGraph() HTML sans %q", want)
		}
	}
}

func TestKBHandlerDetailShowsRelatedKBs(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0002", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0003"); err != nil {
		t.Fatal(err)
	}
	h := NewKBHandler(repo)

	req := httptest.NewRequest("GET", "/kb/detail?id=KB-0001", nil)
	w := httptest.NewRecorder()
	h.Detail(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "/kb/detail?id=KB-0002") {
		t.Errorf("Detail() sans lien vers KB liée KB-0002")
	}
	if !strings.Contains(body, ">kpi<") {
		t.Errorf("Detail() sans badge tag kpi")
	}

	// KB sans tag : section "KB liées" vide
	req = httptest.NewRequest("GET", "/kb/detail?id=KB-0003", nil)
	w = httptest.NewRecorder()
	h.Detail(w, req)
	if !strings.Contains(w.Body.String(), "Aucune KB liée") {
		t.Errorf("Detail() devrait signaler l'absence de KB liée")
	}
}
