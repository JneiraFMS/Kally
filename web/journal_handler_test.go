package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"grz/internal/application/usecases"
	domainjournal "grz/internal/domain/journal"
)

// fakeJournalRepo est un double simple du port de persistance du journal.
type fakeJournalRepo struct {
	saved map[string]domainjournal.Entry
}

func newFakeJournalRepo() *fakeJournalRepo {
	return &fakeJournalRepo{saved: map[string]domainjournal.Entry{}}
}

func (f *fakeJournalRepo) Save(_ context.Context, entry domainjournal.Entry) error {
	f.saved[entry.Date] = entry
	return nil
}

func (f *fakeJournalRepo) Exists(_ context.Context, date string) (bool, error) {
	_, ok := f.saved[date]
	return ok, nil
}

func (f *fakeJournalRepo) ByDate(_ context.Context, date string) (*domainjournal.Entry, error) {
	if entry, ok := f.saved[date]; ok {
		return &entry, nil
	}
	return nil, fmt.Errorf("journal: %w: %s", domainjournal.ErrNotFound, date)
}

func (f *fakeJournalRepo) Entries(_ context.Context) ([]domainjournal.Entry, error) {
	entries := make([]domainjournal.Entry, 0, len(f.saved))
	for _, entry := range f.saved {
		entries = append(entries, entry)
	}
	return entries, nil
}

func (f *fakeJournalRepo) Delete(_ context.Context, date string) error {
	if _, ok := f.saved[date]; !ok {
		return fmt.Errorf("journal: %w: %s", domainjournal.ErrNotFound, date)
	}
	delete(f.saved, date)
	return nil
}

func seedJournalEntry(t *testing.T, repo *fakeJournalRepo, date, content string) {
	t.Helper()
	entry, err := domainjournal.New(date, content)
	if err != nil {
		t.Fatalf("New(%q): %v", date, err)
	}
	if err := repo.Save(context.Background(), entry); err != nil {
		t.Fatalf("Save(%q): %v", date, err)
	}
}

// newTestJournalHandler injecte l'application journal (use cases réels
// sur le fake) dans l'adaptateur HTTP.
func newTestJournalHandler(repo *fakeJournalRepo) *JournalHandler {
	return NewJournalHandler(usecases.NewJournalService(repo))
}

// journalMux câble les routes réelles du journal : les handlers à
// PathValue ({date}) exigent un routage réel (rchooser), les tests suivent
// donc exactement main.go.
func journalMux(repo *fakeJournalRepo) *http.ServeMux {
	h := newTestJournalHandler(repo)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /journal", h.Page)
	mux.HandleFunc("GET /journal/new", h.ShowNewForm)
	mux.HandleFunc("POST /journal", h.Create)
	mux.HandleFunc("GET /journal/{date}", h.ShowEntry)
	mux.HandleFunc("GET /journal/{date}/edit", h.ShowEditForm)
	mux.HandleFunc("POST /journal/edit", h.UpdateEntry)
	mux.HandleFunc("POST /journal/delete", h.Delete)
	return mux
}

func get(t *testing.T, mux *http.ServeMux, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", target, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func post(t *testing.T, mux *http.ServeMux, target string, hx bool, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hx {
		req.Header.Set("HX-Request", "true")
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// TestJournalHandlerPageRendersEntriesAndToday : la page journal liste les
// entrées (les plus récentes d'abord) et propose de créer la note du jour
// quand elle n'existe pas ; quand elle existe, elle est ouvrable.
func TestJournalHandlerPageRendersEntriesAndToday(t *testing.T) {
	repo := newFakeJournalRepo()
	today := journalToday()
	seedJournalEntry(t, repo, "2026-01-01", "panneau de substitution")
	mux := journalMux(repo)

	body := get(t, mux, "/journal").Body.String()
	if !strings.Contains(body, "/journal/2026-01-01") {
		t.Errorf("Page() sans entrée du journal: %s", body)
	}
	if !strings.Contains(body, "Créer la note du jour") {
		t.Errorf("Page() sans bouton de création de la note du jour")
	}
	if strings.Contains(body, "Ouvrir la note du jour") {
		t.Errorf("Page() propose d'ouvrir une note inexistante")
	}

	seedJournalEntry(t, repo, today, "note dejà creee")
	body = get(t, mux, "/journal").Body.String()
	if !strings.Contains(body, "/journal/"+today) {
		t.Errorf("Page() sans lien vers la note du jour existante")
	}
	if strings.Contains(body, "Créer la note du jour") {
		t.Errorf("Page() propose encore de créer la note du jour")
	}
}

// TestJournalHandlerCreateSavesAndRedirects : POST /journal enregistre la
// note libre datée puis ramène sur le journal (303 natif, HX-Redirect htmx).
func TestJournalHandlerCreateSavesAndRedirects(t *testing.T) {
	repo := newFakeJournalRepo()
	mux := journalMux(repo)

	form := url.Values{"date": {"2026-11-03"}, "content": {"**note libre** du jour"}}
	w := post(t, mux, "/journal", true, form)

	if w.Code != 200 {
		t.Errorf("Create(htmx) status = %d, want 200", w.Code)
	}
	if got := w.Header().Get("HX-Redirect"); got != "/journal" {
		t.Errorf("Create(htmx) HX-Redirect = %q, want /journal", got)
	}
	if saved := repo.saved["2026-11-03"].Content; saved != "**note libre** du jour" {
		t.Errorf("Create() content = %q, want conservé tel quel", saved)
	}

	// Flux natif : 303 vers le journal (repo neuf : couvrir le même jour
	// avec le même repo serait un doublon voulu)
	repo2 := newFakeJournalRepo()
	w = post(t, journalMux(repo2), "/journal", false, form)
	if w.Code != http.StatusSeeOther {
		t.Errorf("Create() natif status = %d, want 303", w.Code)
	}
	if _, ok := repo2.saved["2026-11-03"]; !ok {
		t.Errorf("Create() natif n'a pas sauvegardé l'entrée")
	}
}

// TestJournalHandlerCreateRefusesDuplicateDay : créer une deuxième note
// pour un jour couvert est refusé avec un lien vers l'entrée existante,
// sans écraser le contenu original.
func TestJournalHandlerCreateRefusesDuplicateDay(t *testing.T) {
	repo := newFakeJournalRepo()
	seedJournalEntry(t, repo, "2026-11-03", "contenu original")
	mux := journalMux(repo)

	w := post(t, mux, "/journal", false, url.Values{"date": {"2026-11-03"}, "content": {"ÉCRASE"}})

	if w.Code != http.StatusConflict {
		t.Errorf("Create(doublon) status = %d, want 409", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "déjà") {
		t.Errorf("Create(doublon) sans message: %s", body)
	}
	if !strings.Contains(body, "/journal/2026-11-03") {
		t.Errorf("Create(doublon) sans lien vers l'entrée existante")
	}
	if saved := repo.saved["2026-11-03"].Content; saved != "contenu original" {
		t.Errorf("Create(doublon) a écrasé le contenu: %q", saved)
	}
}

// TestJournalHandlerCreateValidation : date invalide ou contenu vierge
// conservés, champ fautif signalé (classe thémée, jamais couleur inline).
func TestJournalHandlerCreateValidation(t *testing.T) {
	repo := newFakeJournalRepo()
	mux := journalMux(repo)

	tests := []struct {
		name     string
		form     url.Values
		wantCode int
		want     string
	}{
		{name: "date invalide", form: url.Values{"date": {"2026-02-30"}, "content": {"x"}}, wantCode: http.StatusBadRequest, want: "field-date"},
		{name: "contenu vide", form: url.Values{"date": {"2026-11-03"}, "content": {"   "}}, wantCode: http.StatusBadRequest, want: "field-content"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := post(t, mux, "/journal", false, tt.form)

			if w.Code != tt.wantCode {
				t.Errorf("Create(%s) status = %d, want %d", tt.name, w.Code, tt.wantCode)
			}
			body := w.Body.String()
			if !strings.Contains(body, tt.want) || !strings.Contains(body, "field-error") {
				t.Errorf("Create(%s) sans signal du champ fautif %q", tt.name, tt.want)
			}
			// Les valeurs saisies sont conservées (flux sans JavaScript)
			if !strings.Contains(body, `name="content"`) {
				t.Errorf("Create(%s) sans formulaire ré-affiché", tt.name)
			}
		})
	}
}

// TestJournalHandlerShowEntryRendersMarkdownAndNavigation : la fiche d'une
// note rend son markdown, la date en clair et la navigation
// précédent/suivant ; une note absente retourne 404.
func TestJournalHandlerShowEntryRendersMarkdownAndNavigation(t *testing.T) {
	repo := newFakeJournalRepo()
	seedJournalEntry(t, repo, "2026-11-01", "ancienne")
	seedJournalEntry(t, repo, "2026-11-03", "note avec **gras**")
	mux := journalMux(repo)

	w := get(t, mux, "/journal/2026-11-03")
	if w.Code != 200 {
		t.Fatalf("ShowEntry() status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{
		"3 novembre 2026",            // la date en clair (Logseq)
		"<strong>gras</strong>",      // le markdown libre est rendu
		`href="/journal/2026-11-01"`, // le jour précédent est navigable
		"/journal/2026-11-03/edit",   // l'édition est proposée
	} {
		if !strings.Contains(body, want) {
			t.Errorf("ShowEntry() sans %q", want)
		}
	}

	if code := get(t, mux, "/journal/2026-10-30").Code; code != http.StatusNotFound {
		t.Errorf("ShowEntry(absente) status = %d, want 404", code)
	}
}

// TestJournalHandlerEditFormPrefilledAndSave : le formulaire d'édition est
// prérempli (date figée, contenu éditable) ; l'enregistrement conserve la
// date et repart vers sa fiche.
func TestJournalHandlerEditFormPrefilledAndSave(t *testing.T) {
	repo := newFakeJournalRepo()
	seedJournalEntry(t, repo, "2026-11-03", "contenu d'origine")
	mux := journalMux(repo)

	w := get(t, mux, "/journal/2026-11-03/edit")
	if w.Code != 200 {
		t.Fatalf("ShowEditForm() status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `value="2026-11-03"`) || !strings.Contains(body, `readonly`) {
		t.Errorf("ShowEditForm() sans date figée: %s", body)
	}
	if !strings.Contains(body, "contenu d&#39;origine") {
		t.Errorf("ShowEditForm() sans contenu prérempli: %s", body)
	}

	w = post(t, mux, "/journal/edit", false, url.Values{"date": {"2026-11-03"}, "content": {"contenu modifié"}})
	if w.Code != http.StatusSeeOther {
		t.Errorf("UpdateEntry() status = %d, want 303", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/journal/2026-11-03" {
		t.Errorf("UpdateEntry() Location = %q, want /journal/2026-11-03", loc)
	}
	if got := repo.saved["2026-11-03"].Content; got != "contenu modifié" {
		t.Errorf("UpdateEntry() content = %q, want contenu modifié", got)
	}
}

// TestJournalHandlerDeleteRemovesEntry : POST /journal/delete retire
// l'entrée et ramène au journal ; absente → 404.
func TestJournalHandlerDeleteRemovesEntry(t *testing.T) {
	repo := newFakeJournalRepo()
	seedJournalEntry(t, repo, "2026-11-03", "x")
	mux := journalMux(repo)

	w := post(t, mux, "/journal/delete", false, url.Values{"date": {"2026-11-03"}})

	if w.Code != http.StatusSeeOther {
		t.Errorf("Delete() status = %d, want 303", w.Code)
	}
	if _, ok := repo.saved["2026-11-03"]; ok {
		t.Errorf("Delete() a laissé l'entrée")
	}

	// Absente : 404 avec message
	w = get(t, mux, "/journal/2026-11-03")
	if w.Code != http.StatusNotFound {
		t.Errorf("ShowEntry(supprimée) status = %d, want 404", w.Code)
	}
}

// TestJournalHandlerErrorPagesEscapeInput : une date frauduleuse passée en
// URL n'apparaît jamais comme HTML brut dans le message d'erreur.
func TestJournalHandlerErrorPagesEscapeInput(t *testing.T) {
	repo := newFakeJournalRepo()
	mux := journalMux(repo)

	payload := `<img src=x onerror=alert(1)>`
	w := get(t, mux, "/journal/"+url.PathEscape(payload))

	body := w.Body.String()
	if strings.Contains(body, "<img ") {
		t.Errorf("ShowEntry() injecte du HTML brut dans le message 404")
	}
	if !strings.Contains(body, "&lt;img ") {
		t.Errorf("ShowEntry() n'échappe pas le texte du message")
	}
}

// TestJournalHandlerBlockedJournals : la créa réussie d'une note proposée
// "`?date=` pour rattraper" (backfill) préremplit le formulaire avec cette date.
func TestJournalHandlerShowNewFormBackfill(t *testing.T) {
	repo := newFakeJournalRepo()
	mux := journalMux(repo)

	w := get(t, mux, "/journal/new?date=2026-10-30")
	if w.Code != 200 {
		t.Fatalf("ShowNewForm() status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, `value="2026-10-30"`) {
		t.Errorf("ShowNewForm(backfill) date non préremplie: %s", body)
	}

	// Une date invalide en query ne doit pas polluer le formulaire
	q := url.Values{"date": {"handwritten garbage"}}
	w = get(t, mux, "/journal/new?"+q.Encode())
	if w.Code != 200 {
		t.Fatalf("ShowNewForm(date invalide) status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); strings.Contains(body, "handwritten garbage") {
		t.Errorf("ShowNewForm() date invalide propagée dans la page")
	}
}
