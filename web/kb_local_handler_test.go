package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// POST /kb/local/serialize : validation du domaine (mêmes règles qu'une KB
// partagée), puis contenu .md canonique renvoyé au script grz-local, qui
// l'écrit dans le dossier local de l'utilisateur. Aucune persistance
// serveur n'est impliquée.

func TestKBLocalHandlerSerializeReturnsMarkdown(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	form := url.Values{
		"id":         {"KB-1"},
		"erreur":     {"trace"},
		"notes":      {"note"},
		"resolution": {"fix"},
		"tags":       {"incident"},
	}
	req := httptest.NewRequest("POST", "/kb/local/serialize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Serialize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Serialize() status = %d (body: %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "---\ntags: [incident]\n---\n\n# KB-1\n") {
		t.Errorf("Serialize() =\n%q", body)
	}
	if w.Header().Get("Content-Type") != "text/markdown; charset=utf-8" {
		t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
}

func TestKBLocalHandlerSerializeRejectsInvalidID(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	form := url.Values{"id": {"../bad"}, "notes": {"note"}}
	req := httptest.NewRequest("POST", "/kb/local/serialize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Serialize(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Serialize(id invalide) status = %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "# KB") {
		t.Errorf("contenu markdown n'aurait jamais dû sortir: %q", w.Body.String())
	}
}

// POST /kb/local/parse : un fichier lu dans le dossier local est renvoyé en
// JSON (champs pour le formulaire et la carte). Corps vide = 400.

func TestKBLocalHandlerParseReturnsFields(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	raw := "---\ntags: [mm]\n---\n\n## Erreur\ntrace\n\n## Notes\nnote /nME22N\n\n## Résolution\nfix\n"
	req := httptest.NewRequest("POST", "/kb/local/parse", strings.NewReader(raw))
	req.Header.Set("Content-Type", "text/markdown; charset=utf-8")
	w := httptest.NewRecorder()

	h.Parse(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Parse() status = %d (body: %s)", w.Code, w.Body.String())
	}
	var parsed struct {
		Erreur     string   `json:"erreur"`
		Notes      string   `json:"notes"`
		Resolution string   `json:"resolution"`
		Tags       []string `json:"tags"`
		Module     string   `json:"module"`
		TCodes     []string `json:"tcodes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("Parse() JSON: %v", err)
	}
	if parsed.Notes != "note /nME22N" || parsed.Resolution != "fix" || parsed.Tags[0] != "mm" {
		t.Errorf("Parse() = %+v", parsed)
	}
	// Dérivations : mêmes règles que les cartes serveur (aucune
	// duplication côté client).
	if parsed.Module != "SAP MM" {
		t.Errorf("Parse() module = %q, want SAP MM", parsed.Module)
	}
	if len(parsed.TCodes) != 1 || parsed.TCodes[0] != "ME22N" {
		t.Errorf("Parse() tcodes = %v, want [ME22N]", parsed.TCodes)
	}
}

func TestKBLocalHandlerParseRejectsEmptyBody(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	req := httptest.NewRequest("POST", "/kb/local/parse", strings.NewReader(""))
	w := httptest.NewRecorder()

	h.Parse(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Parse(corps vide) status = %d", w.Code)
	}
}

// POST /kb/local/render : contrat JSON {notes, resolution} → fragments HTML
// rendus avec les IDs partagés connus — un [[ID]] existant pointe vers la
// fiche, une cible absente est un lien mort (lien de création prérempli).
// L'Erreur n'emprunte jamais ce pipeline : trace brute côté client.

func TestKBLocalHandlerRenderResolvesAgainstSharedIDs(t *testing.T) {
	repo := newFakeRepo()
	repo.saved["KB-2"] = domainKB("KB-2")
	h := newTestLocalHandler(repo)
	body := `{"notes":"voir [[KB-2]]","resolution":"et [[KB-X]]"}`
	req := httptest.NewRequest("POST", "/kb/local/render", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Render(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Render() status = %d (body: %s)", w.Code, w.Body.String())
	}
	var response struct {
		NotesHTML      string `json:"notesHTML"`
		ResolutionHTML string `json:"resolutionHTML"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Render() JSON: %v", err)
	}
	if !strings.Contains(response.NotesHTML, `href="/kb/detail?id=KB-2"`) {
		t.Errorf("lien vers KB partagée attendu: %q", response.NotesHTML)
	}
	if !strings.Contains(response.ResolutionHTML, "wikilink-missing") {
		t.Errorf("lien mort attendu pour une cible partagée absente: %q", response.ResolutionHTML)
	}
}

func TestKBLocalHandlerRenderRejectsEmptyBody(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	req := httptest.NewRequest("POST", "/kb/local/render", strings.NewReader(`{"notes":"","resolution":""}`))
	w := httptest.NewRecorder()

	h.Render(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Render(contenu vide) status = %d", w.Code)
	}
}

func TestKBLocalHandlerRenderRejectsBadJSON(t *testing.T) {
	h := newTestLocalHandler(newFakeRepo())
	req := httptest.NewRequest("POST", "/kb/local/render", strings.NewReader("pas du json"))
	w := httptest.NewRecorder()

	h.Render(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Render(json invalide) status = %d", w.Code)
	}
}
