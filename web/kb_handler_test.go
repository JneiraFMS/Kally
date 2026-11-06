package web

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"grz/internal/application/usecases"
	domainkb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
)

// fakeRepo est un double simple du application repository port pour les tests.
type fakeRepo struct {
	saved map[string]domainkb.KB
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{saved: map[string]domainkb.KB{}}
}

func (f *fakeRepo) Save(_ context.Context, k domainkb.KB) error {
	f.saved[k.ID] = k
	return nil
}

func (f *fakeRepo) Exists(_ context.Context, id string) (bool, error) {
	_, ok := f.saved[id]
	return ok, nil
}

func (f *fakeRepo) FindAll(_ context.Context) ([]domainkb.KB, error) {
	kbs := []domainkb.KB{}
	for _, kb := range f.saved {
		kbs = append(kbs, kb)
	}
	return kbs, nil
}

func (f *fakeRepo) FindByID(_ context.Context, id string) (*domainkb.KB, error) {
	if b, ok := f.saved[id]; ok {
		return &b, nil
	}
	return nil, fmt.Errorf("kb: %w: %s", domainkb.ErrNotFound, id)
}

func (f *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := f.saved[id]; !ok {
		return fmt.Errorf("kb: %w: %s", domainkb.ErrNotFound, id)
	}
	delete(f.saved, id)
	return nil
}

func newTestHandler(repo *fakeRepo) *KBHandler {
	return NewKBHandler(usecases.NewKBService(repo))
}

// newTestLocalHandler construit l'adaptateur HTTP de la bibliothèque locale
// au-dessus du même service KB.
func newTestLocalHandler(repo *fakeRepo) *KBLocalHandler {
	return NewKBLocalHandler(usecases.NewKBService(repo))
}

func TestKBHandlerCreateSavesKB(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	form := url.Values{
		"id":         {"KB-0001"},
		"erreur":     {"Div par zéro"},
		"notes":      {"Tester le diviseur"},
		"resolution": {"Vérifier avant division"},
	}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true") // formulaire htmx
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != 200 {
		t.Errorf("Create() status = %d, want 200", w.Code)
	}
	if _, ok := repo.saved["KB-0001"]; !ok {
		t.Errorf("Create() n'a pas sauvegardé la KB, saved = %v", repo.saved)
	}
	// Après enregistrement : navigation vers la fiche (htmx reçoit l'ordre
	// via l'en-tête HX-Redirect).
	if loc := w.Header().Get("HX-Redirect"); loc != "/kb/detail?id=KB-0001" {
		t.Errorf("Create() HX-Redirect = %q, want /kb/detail?id=KB-0001", loc)
	}
}

// TestKBHandlerCreateRedirectsNativeFlow : sans htmx (formulaire natif),
// l'enregistrement se termine par une redirection 303 vers la fiche.
func TestKBHandlerCreateRedirectsNativeFlow(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}, "erreur": {"e"}, "notes": {"n"}, "resolution": {"r"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Create() status = %d, want 303", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/kb/detail?id=KB-0001" {
		t.Errorf("Create() Location = %q, want /kb/detail?id=KB-0001", loc)
	}
}

// TestKBHandlerCreateRefusesDuplicateID : créer avec un ID existant est
// refusé (le Save n'écrase pas) avec un message suggérant la modification.
func TestKBHandlerCreateRefusesDuplicateID(t *testing.T) {
	repo := seedRepo(t, "KB-0001")
	original := repo.saved["KB-0001"]
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}, "erreur": {"NOUVEAU CONTENU"}, "notes": {"n"}, "resolution": {"r"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Create() status = %d, want 409", w.Code)
	}
	if got := repo.saved["KB-0001"]; got.Erreur != original.Erreur {
		t.Errorf("Create() a écrasé la KB existante: %q → %q", original.Erreur, got.Erreur)
	}
	body := w.Body.String()
	if !strings.Contains(body, "existe déjà") {
		t.Errorf("Create() sans message de doublon: %s", body)
	}
	if !strings.Contains(body, "/kb/edit?id=KB-0001") {
		t.Errorf("Create() sans lien vers la modification: %s", body)
	}
}

// TestKBHandlerErrorPagesEscapeInput : les IDs non validés passés en query
// ne doivent jamais apparaître comme HTML brut dans les pages d'erreur.
func TestKBHandlerErrorPagesEscapeInput(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	payload := `<img src=x onerror=alert(1)>`

	req := httptest.NewRequest("GET", "/kb/detail?id="+url.QueryEscape(payload), nil)
	w := httptest.NewRecorder()
	h.Detail(w, req)
	body := w.Body.String()
	if strings.Contains(body, "<img ") {
		t.Errorf("Detail() injecte du HTML brut dans le message 404: %s", body)
	}
	if !strings.Contains(body, "&lt;img src=x onerror=alert(1)&gt;") {
		t.Errorf("Detail() n'échappe pas le texte du message: %s", body)
	}

	// Même exigence sur la page d'édition introuvable
	req = httptest.NewRequest("GET", "/kb/edit?id="+url.QueryEscape(payload), nil)
	w = httptest.NewRecorder()
	h.ShowEditForm(w, req)
	if strings.Contains(w.Body.String(), "<img ") {
		t.Errorf("ShowEditForm() injecte du HTML brut dans le message 404")
	}
}

// TestKBHandlerCreatePreservesTagsValidation : un tag invalide reste rejeté.
func TestKBHandlerCreatePreservesTagsValidation(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}, "erreur": {"e"}, "notes": {"n"}, "resolution": {"r"}, "tags": {"bad tag!"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() status = %d, want 400", w.Code)
	}
	if len(repo.saved) != 0 {
		t.Errorf("Create() a sauvegardé une KB invalide: %v", repo.saved)
	}
}

// TestKBHandlerDeleteRemovesKB : POST /kb/delete supprime la KB et ramène
// sur l'accueil.
func TestKBHandlerDeleteRemovesKB(t *testing.T) {
	repo := seedRepo(t, "KB-0001", "KB-0002")
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}}
	req := httptest.NewRequest("POST", "/kb/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("Delete() status = %d, want 303", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "/" {
		t.Errorf("Delete() Location = %q, want /", loc)
	}
	if _, ok := repo.saved["KB-0001"]; ok {
		t.Errorf("Delete() KB-0001 toujours présent: %v", repo.saved)
	}
	if _, ok := repo.saved["KB-0002"]; !ok {
		t.Errorf("Delete() a supprimé une autre KB: %v", repo.saved)
	}
}

// TestKBHandlerDeleteRedirectsHTMX : sous htmx, la suppression doit répondre
// 200 + HX-Redirect. htmx suit une 303 et change la cible dans le DOM : le
// navigateur resterait sur l'URL de la KB supprimée au lieu de naviguer.
func TestKBHandlerDeleteRedirectsHTMX(t *testing.T) {
	repo := seedRepo(t, "KB-0001", "KB-0002")
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}}
	req := httptest.NewRequest("POST", "/kb/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Delete() status = %d, want 200 (htmx suit la 303 et reste sur la KB supprimée)", w.Code)
	}
	if got := w.Header().Get("HX-Redirect"); got != "/" {
		t.Errorf("Delete() HX-Redirect = %q, want /", got)
	}
	if loc := w.Header().Get("Location"); loc != "" {
		t.Errorf("Delete() Location = %q sous htmx, want vide", loc)
	}
	if _, ok := repo.saved["KB-0001"]; ok {
		t.Errorf("Delete() htmx n'a pas supprimé la KB")
	}
	if _, ok := repo.saved["KB-0002"]; !ok {
		t.Errorf("Delete() htmx a supprimé une autre KB")
	}
}

// TestKBHandlerDeleteMissingKB : supprimer une KB absente répond 404.
func TestKBHandlerDeleteMissingKB(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	form := url.Values{"id": {"KB-INCONNU"}}
	req := httptest.NewRequest("POST", "/kb/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Delete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Delete() status = %d, want 404", w.Code)
	}
}

func TestKBHandlerCreateRejectsInvalidID(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

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

func TestKBHandlerCreateRejectsMalformedForm(t *testing.T) {
	h := newTestHandler(newFakeRepo())
	req := httptest.NewRequest("POST", "/kb", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() status = %d, want 400", w.Code)
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
	k, err := domainkb.NewKB(id, "erreur démo "+id, "notes", "résolution", tags...)
	if err != nil {
		t.Fatalf("NewKB(%q): %v", id, err)
	}
	return repo.Save(context.Background(), k)
}

func TestKBHandlerHomeShowsKBCards(t *testing.T) {
	repo := seedRepo(t, "KB-0001", "KB-0002")
	h := newTestHandler(repo)

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
	h := newTestHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)

	if !strings.Contains(w.Body.String(), "Aucune KB") {
		t.Errorf("Home() ne gère pas l'état vide")
	}
}

func TestKBHandlerCreateSavesTags(t *testing.T) {
	repo := newFakeRepo()
	h := newTestHandler(repo)

	form := url.Values{
		"id":         {"KB-0001"},
		"erreur":     {"e"},
		"notes":      {"n"},
		"resolution": {"r"},
		"tags":       {" Incident, KPI "},
	}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true") // formulaire htmx
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

// TestKBHandlerHomeFiltersByTag : le filtre ?tag= vit sur l'accueil
// (la page liste a été fusionnée avec l'accueil).
func TestKBHandlerHomeFiltersByTag(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0002", "csv"); err != nil {
		t.Fatal(err)
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest("GET", "/?tag=kpi", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "KB-0001") {
		t.Errorf("Home(filter kpi) n'affiche pas KB-0001")
	}
	if strings.Contains(body, "KB-0002") {
		t.Errorf("Home(filter kpi) affiche KB-0002 hors filtre")
	}
	if !strings.Contains(body, "aria-pressed=") {
		t.Errorf("Home(filter kpi) : le tag de filtre actif doit être un bouton segmenté (aria-pressed)")
	}

	// Sans filtre : toutes les KB
	req = httptest.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	h.Home(w, req)
	if !strings.Contains(w.Body.String(), "KB-0002") {
		t.Errorf("Home(sans filtre) doit afficher toutes les KB")
	}
}

// TestKBHandlerListRedirectsToHome : /kb redirige vers l'accueil
// (avec conservation du filtre ?tag= éventuel).
func TestKBHandlerListRedirectsToHome(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	tests := []struct {
		name     string
		target   string
		location string
	}{
		{name: "sans filtre", target: "/kb", location: "/"},
		{name: "avec filtre", target: "/kb?tag=kpi", location: "/?tag=kpi"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.target, nil)
			w := httptest.NewRecorder()
			h.List(w, req)

			if w.Code != http.StatusSeeOther {
				t.Errorf("List(%s) status = %d, want 303", tt.target, w.Code)
			}
			if loc := w.Header().Get("Location"); loc != tt.location {
				t.Errorf("List(%s) Location = %q, want %q", tt.target, loc, tt.location)
			}
		})
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
	h := newTestHandler(repo)

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

// TestKBHandlerShowNewFormPrefillsID : un lien mort [[KB-0042]] ouvre le
// formulaire de création avec l'identifiant déjà saisi.
func TestKBHandlerShowNewFormPrefillsID(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/kb/new?id=KB-0042", nil)
	w := httptest.NewRecorder()
	h.ShowNewForm(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ShowNewForm() status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, `value="KB-0042"`) {
		t.Errorf("ShowNewForm() ne préremplit pas l'identifiant: %s", body)
	}
}

// TestKBHandlerShowNewFormIgnoresInvalidID : un identifiant invalide n'est
// ni prérempli ni répercuté dans la page (aucune injection).
func TestKBHandlerShowNewFormIgnoresInvalidID(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	req := httptest.NewRequest("GET", "/kb/new?id="+url.QueryEscape("../hack"), nil)
	w := httptest.NewRecorder()
	h.ShowNewForm(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ShowNewForm() status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); strings.Contains(body, "../hack") {
		t.Errorf("ShowNewForm() répercute un identifiant invalide: %s", body)
	}
}

// TestKBHandlerDetailRendersWikiLinksAndBacklinks : de bout en bout, la fiche
// rend les liens [[...]] (existant ou mort) et les rétroliens orientés.
func TestKBHandlerDetailRendersWikiLinksAndBacklinks(t *testing.T) {
	repo := newFakeRepo()
	source, err := domainkb.NewKB("KB-0001", "erreur source", "voir [[KB-0002]] et [[KB-9999]]", "")
	if err != nil {
		t.Fatalf("NewKB(source): %v", err)
	}
	target, err := domainkb.NewKB("KB-0002", "erreur cible", "", "")
	if err != nil {
		t.Fatalf("NewKB(target): %v", err)
	}
	if err := repo.Save(context.Background(), source); err != nil {
		t.Fatalf("Save(source): %v", err)
	}
	if err := repo.Save(context.Background(), target); err != nil {
		t.Fatalf("Save(target): %v", err)
	}
	h := newTestHandler(repo)

	// La source rend un lien existant et un lien mort.
	req := httptest.NewRequest("GET", "/kb/detail?id=KB-0001", nil)
	w := httptest.NewRecorder()
	h.Detail(w, req)
	body := w.Body.String()
	if !strings.Contains(body, `href="/kb/detail?id=KB-0002"`) {
		t.Errorf("Detail(KB-0001) sans lien wiki vers KB-0002")
	}
	if !strings.Contains(body, `href="/kb/new?id=KB-9999"`) {
		t.Errorf("Detail(KB-0001) sans lien mort vers la création de KB-9999")
	}

	// La cible affiche la source en rétrolien, sans réciprocité.
	req = httptest.NewRequest("GET", "/kb/detail?id=KB-0002", nil)
	w = httptest.NewRecorder()
	h.Detail(w, req)
	body = w.Body.String()
	if !strings.Contains(body, "Références liées") {
		t.Errorf("Detail(KB-0002) sans section de rétroliens")
	}
	if !strings.Contains(body, "erreur source") {
		t.Errorf("Detail(KB-0002) sans extrait du rétrolien KB-0001")
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
	h := newTestHandler(repo)

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

// htmx ne remplace pas la cible d'une réponse 4xx/5xx : une erreur de
// validation doit donc partir en 200 avec le message seul, sinon l'échange
// ne montre rien à l'utilisateur.
func TestKBHandlerCreateInvalidHTMXSwapsMessage(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	form := url.Values{"id": {"../hack"}, "erreur": {"e"}, "notes": {"n"}, "resolution": {"r"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Create() status = %d, want 200 (htmx ignore les 4xx)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "identifiant invalide") {
		t.Errorf("Create() sans message de validation: %s", body)
	}
	if !strings.Contains(body, "field-error") {
		t.Errorf("Create() sans signalement du champ fautif: %s", body)
	}
	if strings.Contains(body, "<html") {
		t.Errorf("Create() renvoie une page complète dans #reponse-serveur")
	}
}

// TestKBHandlerCreateDuplicateHTMXSwapsMessage : le doublon suit la même
// règle (200 + message) pour être visible sous le formulaire.
func TestKBHandlerCreateDuplicateHTMXSwapsMessage(t *testing.T) {
	repo := seedRepo(t, "KB-0001")
	h := newTestHandler(repo)

	form := url.Values{"id": {"KB-0001"}, "erreur": {"e"}, "notes": {"n"}, "resolution": {"r"}}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Create(doublon) status = %d, want 200 (htmx ignore les 4xx)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "existe déjà") || !strings.Contains(body, "/kb/edit?id=KB-0001") {
		t.Errorf("Create(doublon) sans message ni lien vers la modification: %s", body)
	}
}

// TestKBHandlerCreateInvalidNativeKeepsValues : sans JavaScript, la page
// complète est renvoyée avec le message et les valeurs déjà saisies.
func TestKBHandlerCreateInvalidNativeKeepsValues(t *testing.T) {
	h := newTestHandler(newFakeRepo())

	form := url.Values{
		"id":         {"../hack"},
		"erreur":     {"trace conservee"},
		"notes":      {"notes saisies par l'utilisateur"},
		"resolution": {"r"},
	}
	req := httptest.NewRequest("POST", "/kb", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Create() status = %d, want 400 (flux natif)", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "identifiant invalide") {
		t.Errorf("Create() sans message de validation: %s", body)
	}
	if !strings.Contains(body, "notes saisies par l&#39;utilisateur") &&
		!strings.Contains(body, "notes saisies par l'utilisateur") {
		t.Errorf("Create() ne conserve pas les notes saisies: %s", body)
	}
	if !strings.Contains(body, `id="reponse-serveur"`) {
		t.Errorf("Create() sans zone de réponse dans la page renvoyée")
	}
}

// TestKBHandlerUpdateInvalidKeepsValues : même exigence sur le formulaire
// d'édition (valeurs renvoyées en natif, message seul en htmx).
func TestKBHandlerUpdateInvalidKeepsValues(t *testing.T) {
	repo := seedRepo(t, "KB-0001")
	h := newTestHandler(repo)

	form := url.Values{
		"id":         {"KB-0001"},
		"erreur":     {"e"},
		"notes":      {"notes modifiees"},
		"resolution": {"r"},
		"tags":       {"bad tag!"},
	}

	// Flux natif : page complète, valeurs conservées, 400.
	req := httptest.NewRequest("POST", "/kb/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.UpdateForm(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("UpdateForm() status = %d, want 400", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "notes modifiees") {
		t.Errorf("UpdateForm() ne conserve pas les notes modifiées: %s", body)
	}
	if !strings.Contains(body, "invalide") {
		t.Errorf("UpdateForm() sans message de validation: %s", body)
	}

	// Flux htmx : 200 + message seul.
	req = httptest.NewRequest("POST", "/kb/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	w = httptest.NewRecorder()
	h.UpdateForm(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("UpdateForm(htmx) status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "invalide") || strings.Contains(body, "<html") {
		t.Errorf("UpdateForm(htmx) doit renvoyer le message seul: %s", body)
	}
}

// TestKBHandlerValidationFocusUsesThemeClass : le champ fautif est signalé
// par une classe (thème clair/sombre) et non par un contour inline dont la
// couleur est codée en dur.
func TestKBHandlerValidationFocusUsesThemeClass(t *testing.T) {
	body := renderString(t, validationMessage(domainkb.ErrInvalidID))

	if strings.Contains(body, "#dc322f") {
		t.Errorf("couleur d'erreur codée en dur: %s", body)
	}
	if !strings.Contains(body, "classList.add") || !strings.Contains(body, "field-error") {
		t.Errorf("le champ fautif n'est pas signalé par la classe field-error: %s", body)
	}
}

// TestKBHandlerHomeFiltrePlusieursTags : ?tag=a&tag=b filtre en AND (à la
// Logseq) — seules les KB portant TOUS les tags sont listées.
func TestKBHandlerHomeFiltrePlusieursTags(t *testing.T) {
	repo := newFakeRepo()
	if err := seedKB(t, repo, "KB-0001", "incident", "kpi"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0002", "incident"); err != nil {
		t.Fatal(err)
	}
	if err := seedKB(t, repo, "KB-0003", "kpi"); err != nil {
		t.Fatal(err)
	}
	h := newTestHandler(repo)

	req := httptest.NewRequest("GET", "/?tag=incident&tag=kpi", nil)
	w := httptest.NewRecorder()
	h.Home(w, req)

	if w.Code != 200 {
		t.Errorf("Home(multi tags) status = %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "KB-0001") {
		t.Errorf("Home(multi tags) n'affiche pas KB-0001")
	}
	if strings.Contains(body, "KB-0002") || strings.Contains(body, "KB-0003") {
		t.Errorf("Home(multi tags) affiche une KB hors filtre AND: %s", body)
	}
}

func renderString(t *testing.T, node gomponents.Node) string {
	t.Helper()
	var buf strings.Builder
	if err := node.Render(&buf); err != nil {
		t.Fatalf("Render(): %v", err)
	}
	return buf.String()
}
