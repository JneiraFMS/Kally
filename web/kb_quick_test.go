package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	domainkb "grz/internal/domain/kb"
)

// newQuickRequest construit une requête POST /kb/quick en mode htmx.
func newQuickRequest(form url.Values) *http.Request {
	req := httptest.NewRequest("POST", "/kb/quick", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	return req
}

// TestKBHandlerQuickCreate : la création libre stocke une KB avec l'ID saisi
// à la main, la note libre (images collables) et des sections Erreur et
// Résolution laissées vides (complétables en édition).
func TestKBHandlerQuickCreate(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)
	w := httptest.NewRecorder()

	h.CreateQuick(w, newQuickRequest(url.Values{
		"id":    {"KB-Q1"},
		"notes": {"note libre"},
		"tags":  {"calcul"},
	}))

	if w.Code != http.StatusOK {
		t.Fatalf("CreateQuick() status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if loc := w.Header().Get("HX-Redirect"); loc != "/kb/detail?id=KB-Q1" {
		t.Errorf("CreateQuick() HX-Redirect = %q, want fiche KB-Q1", loc)
	}
	saved, ok := repo.saved["KB-Q1"]
	if !ok {
		t.Fatalf("CreateQuick() n'a pas sauvegardé la KB: %v", repo.saved)
	}
	if saved.Notes != "note libre" {
		t.Errorf("Notes = %q, want %q", saved.Notes, "note libre")
	}
	if saved.Erreur != "" || saved.Resolution != "" {
		t.Errorf("sections non demandées complétées: %+v", saved)
	}
	if len(saved.Tags) != 1 || saved.Tags[0] != "calcul" {
		t.Errorf("Tags = %v, want [calcul]", saved.Tags)
	}
}

// TestKBHandlerQuickCreateDuplicateRerendersQuickForm : un doublon ré-affiche
// le formulaire libre (pas le formulaire complet) avec le message d'erreur.
func TestKBHandlerQuickCreateDuplicateRerendersQuickForm(t *testing.T) {
	repo := newFakeRepo()
	repo.saved["KB-Q1"] = domainKB("KB-Q1")
	h := newTestHandler(repo)
	w := httptest.NewRecorder()

	h.CreateQuick(w, newQuickRequest(url.Values{"id": {"KB-Q1"}, "notes": {"note"}}))

	if w.Code != http.StatusOK {
		t.Errorf("CreateQuick(doublon htmx) status = %d, want 200 (message dans la zone)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "existe déjà") {
		t.Errorf("message de doublon absent:\n%s", body)
	}
	if strings.Contains(body, `id="field-resolution"`) {
		t.Errorf("formulaire complet rendu à la place du formulaire libre:\n%s", body)
	}
}

// TestKBHandlerQuickCreateInvalidIDRerendersQuickForm : un identifiant
// invalide renvoie 400 (flux natif) avec la page du formulaire libre.
func TestKBHandlerQuickCreateInvalidIDRerendersQuickForm(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	req := newQuickRequest(url.Values{"id": {"../bad"}, "notes": {"note"}})
	req.Header.Set("HX-Request", "") // POST natif
	w := httptest.NewRecorder()

	h.CreateQuick(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("CreateQuick(id invalide) status = %d, want 400", w.Code)
	}
	body := w.Body.String()
	if strings.Contains(body, `id="field-resolution"`) {
		t.Errorf("formulaire complet rendu à la place du formulaire libre:\n%s", body)
	}
}

// TestKBHandlerQuickCreateEmptyNotes : la note peut être vide (la fiche
// reste validable et se remplie après).
func TestKBHandlerQuickCreateEmptyNotes(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)
	w := httptest.NewRecorder()

	h.CreateQuick(w, newQuickRequest(url.Values{"id": {"KB-E1"}}))

	if w.Code != http.StatusOK {
		t.Fatalf("CreateQuick() status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if saved, ok := repo.saved["KB-E1"]; !ok || saved.Notes != "" {
		t.Errorf("KB-E1 = %+v, want présente avec notes vides", saved)
	}
}

// domainKB construit une KB valide pour les doubles de tests.
func domainKB(id string) domainkb.KB {
	k, _ := domainkb.NewKB(id, "e", "n", "r", "t")
	return k
}
