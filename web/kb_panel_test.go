package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainkb "grz/internal/domain/kb"
)

// TestKBHandlerPanelFragment : GET /kb/panel?id=… sert un FRAGMENT injecté
// dans le tiroir latéral — jamais une page complète (pas de doctype, pas
// de navbar, pas de feuille rechargée). Le détail reste accessible via le
// lien historique en HTML natif.
func TestKBHandlerPanelFragment(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "kpi"); err != nil {
		t.Fatal(err)
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest("GET", "/kb/panel?id=KB-0001", nil)
	w := httptest.NewRecorder()
	h.Panel(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Panel() status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="kb-panel"`) {
		t.Errorf("fragment sans conteneur kb-panel:\n%.400s", body)
	}
	if !strings.Contains(body, "KB-0001") {
		t.Errorf("fragment sans l'identifiant de la fiche")
	}
	if strings.Contains(body, "<!doctype") || strings.Contains(body, "<nav") {
		t.Errorf("Panel() a rendu une page complète au lieu du fragment")
	}
}

// TestKBHandlerPanelNotFound : fiche inconnue → 404 (le tiroir garde son
// message d'erreur locale via le catch du script).
func TestKBHandlerPanelNotFound(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/kb/panel?id=KB-404", nil)
	w := httptest.NewRecorder()
	h.Panel(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Panel(fiche inconnue) status = %d, want 404", w.Code)
	}
}

// TestKBHandlerPanelInvalidID : identifiant dangereux → 400, jamais une
// lecture de fichier (les tests de l'entité garantissent déjà la
// validation ; ici on sert de garde-fou outlet).
func TestKBHandlerPanelInvalidID(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/kb/panel?id=../secret", nil)
	w := httptest.NewRecorder()
	h.Panel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Panel(id invalide) status = %d, want 400", w.Code)
	}
}

// TestKBHandlerPanelWikiLinks : le fragment calcule aussi les liens wiki +
// rétroliens (les KnownIDs passent par le même use case Detail).
func TestKBHandlerPanelWikiLinks(t *testing.T) {
	repo := newFakeRepo()
	source, err := domainkb.NewKB("KB-0001", "erreur source", "voir [[KB-0002]]", "")
	if err != nil {
		t.Fatalf("NewKB(source): %v", err)
	}
	target, err := domainkb.NewKB("KB-0002", "erreur cible", "", "")
	if err != nil {
		t.Fatalf("NewKB(target): %v", err)
	}
	if err := repo.Save(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), target); err != nil {
		t.Fatal(err)
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest("GET", "/kb/panel?id=KB-0001", nil)
	w := httptest.NewRecorder()
	h.Panel(w, req)

	if !strings.Contains(w.Body.String(), `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("fragment sans lien wiki vers KB-0002 (KnownIDs non propagés ?)")
	}
}
