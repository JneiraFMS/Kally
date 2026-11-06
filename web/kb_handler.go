package web

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"grz/html/pages"
	"grz/internal/application"
	"grz/internal/application/dto"
	"grz/internal/application/ports/inbound"
	domainkb "grz/internal/domain/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// KBHandler is the inbound HTTP adapter for knowledge-base use cases.
type KBHandler struct {
	app inbound.KBApplication
}

// NewKBHandler injects the application port into the HTTP adapter.
func NewKBHandler(app inbound.KBApplication) *KBHandler {
	return &KBHandler{app: app}
}

// Home displays the knowledge-base cards, filtered by the ?tag= values
// (AND when several tags are provided, like a Logseq query).
func (h *KBHandler) Home(w http.ResponseWriter, r *http.Request) {
	tags := domainkb.NormalizeTagFilter(r.URL.Query()["tag"])
	items, err := h.app.List(r.Context(), tags...)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	renderNode(w, pages.PageAccueil(r.URL.Path, items, tags...))
}

// List redirects old list URLs to the home page, keeping the multi-tag filter.
func (h *KBHandler) List(w http.ResponseWriter, r *http.Request) {
	target := "/"
	if tags := r.URL.Query()["tag"]; len(tags) > 0 {
		target = "/?" + url.Values{"tag": tags}.Encode()
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// ShowNewForm displays the create form. A valid ?id= value pre-fills the
// identifier field, so a missing wiki link [[KB-9999]] opens a ready-to-save
// form. An invalid identifier is ignored rather than reflected.
func (h *KBHandler) ShowNewForm(w http.ResponseWriter, r *http.Request) {
	form := pages.FormulaireKB{}
	if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
		if err := domainkb.ValidateID(id); err == nil {
			form.Valeurs = map[string]string{"id": id}
		}
	}
	renderNode(w, pages.PageKBCreateWith(r.URL.Path, form))
}

// ShowNewQuickForm displays the free-form creation: a hand-written ID, a
// single markdown note (pasteable images) and optional tags. Erreur and
// Résolution stay fillable through the full edit form.
func (h *KBHandler) ShowNewQuickForm(w http.ResponseWriter, r *http.Request) {
	renderNode(w, pages.PageKBCreateQuick(r.URL.Path))
}

// CreateQuick validates and saves the free-form creation. It reuses the
// same Create use case; validation errors re-render the free form instead
// of the full one.
func (h *KBHandler) CreateQuick(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	item, err := h.app.Create(r.Context(), inputFromRequest(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrAlreadyExists):
			h.renderQuickFormError(w, r, http.StatusConflict,
				duplicateMessage(strings.TrimSpace(r.FormValue("id"))))
		case errors.Is(err, domainkb.ErrInvalidID), errors.Is(err, domainkb.ErrInvalidTag):
			h.renderQuickFormError(w, r, http.StatusBadRequest, validationMessage(err))
		default:
			writeInternalError(w, err)
		}
		return
	}
	redirectToDetail(w, r, item.ID)
}

// renderQuickFormError répond à une erreur du formulaire libre (page
// complète pré-remplie en natif, message seul en htmx).
func (h *KBHandler) renderQuickFormError(w http.ResponseWriter, r *http.Request, status int, message gomponents.Node) {
	respondFormError(w, r, status, message,
		pages.PageKBCreateQuickWith(r.URL.Path,
			pages.FormulaireKB{Valeurs: formValues(r), Message: message}))
}

// ShowGraph displays the global knowledge-base graph.
func (h *KBHandler) ShowGraph(w http.ResponseWriter, r *http.Request) {
	items, err := h.app.List(r.Context())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	renderNode(w, pages.PageKBGraph(r.URL.Path, items))
}

// Create validates and creates a knowledge base submitted by the form.
func (h *KBHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	item, err := h.app.Create(r.Context(), inputFromRequest(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrAlreadyExists):
			h.renderFormError(w, r, http.StatusConflict, duplicateMessage(strings.TrimSpace(r.FormValue("id"))),
				func(form pages.FormulaireKB) gomponents.Node {
					return pages.PageKBCreateWith(r.URL.Path, form)
				})
		case errors.Is(err, domainkb.ErrInvalidID), errors.Is(err, domainkb.ErrInvalidTag):
			h.renderFormError(w, r, http.StatusBadRequest, validationMessage(err),
				func(form pages.FormulaireKB) gomponents.Node {
					return pages.PageKBCreateWith(r.URL.Path, form)
				})
		default:
			writeInternalError(w, err)
		}
		return
	}
	redirectToDetail(w, r, item.ID)
}

// UpdateForm validates and saves edits to an existing knowledge base.
func (h *KBHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	item, err := h.app.Update(r.Context(), inputFromRequest(r))
	if err != nil {
		if errors.Is(err, domainkb.ErrInvalidID) || errors.Is(err, domainkb.ErrInvalidTag) {
			id := r.FormValue("id")
			h.renderFormError(w, r, http.StatusBadRequest, validationMessage(err),
				func(form pages.FormulaireKB) gomponents.Node {
					return pages.PageKBEditWith(r.URL.Path, id, form)
				})
		} else {
			writeInternalError(w, err)
		}
		return
	}
	redirectToDetail(w, r, item.ID)
}

// respondFormError répond à une erreur de formulaire (formes KB et journal).
//
// htmx n'échange pas le contenu d'une réponse 4xx/5xx : pour un échange htmx
// on renvoie donc 200 + le message seul (la zone #reponse-serveur est
// remplacée). Pour le flux natif (sans JavaScript), on renvoie la page
// complète pré-remplie avec le statut d'erreur approprié.
func respondFormError(w http.ResponseWriter, r *http.Request, status int, message gomponents.Node, page gomponents.Node) {
	if r.Header.Get("HX-Request") == "true" {
		writeMessage(w, http.StatusOK, message)
		return
	}
	writeMessage(w, status, page)
}

// renderFormError répond à une erreur de formulaire KB
// (page complète pré-remplie en natif, message seul en htmx).
func (h *KBHandler) renderFormError(w http.ResponseWriter, r *http.Request, status int, message gomponents.Node, page func(pages.FormulaireKB) gomponents.Node) {
	respondFormError(w, r, status, message, page(pages.FormulaireKB{Valeurs: formValues(r), Message: message}))
}

// formValues collecte les valeurs saisies, indexées par name, pour
// ré-afficher le formulaire sans perdre la saisie.
func formValues(r *http.Request) map[string]string {
	valeurs := make(map[string]string, 5)
	for _, nom := range []string{"id", "erreur", "notes", "resolution", "tags"} {
		valeurs[nom] = r.FormValue(nom)
	}
	return valeurs
}

// ShowEditForm displays the pre-filled edit form.
func (h *KBHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	item, ok := h.fetch(w, r, r.URL.Query().Get("id"))
	if !ok {
		return
	}
	renderNode(w, pages.PageKBEdit(r.URL.Path, *item))
}

// Delete removes a knowledge base submitted by the form.
func (h *KBHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	id := r.FormValue("id")
	if err := h.app.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, domainkb.ErrNotFound):
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
		case errors.Is(err, domainkb.ErrInvalidID):
			writeMessage(w, http.StatusBadRequest, errorMessage("Identifiant de KB invalide"))
		default:
			writeInternalError(w, err)
		}
		return
	}
	redirectTo(w, r, "/")
}

// errorMessage builds a neutral HTML message; gomponents escapes its text.
func errorMessage(text string) gomponents.Node {
	return Span(Class("text-g-negative-text font-semibold"), gomponents.Text(text))
}

// validationMessage identifies the field that failed domain validation.
func validationMessage(err error) gomponents.Node {
	field := "tags"
	text := "Données invalides : " + err.Error()
	if errors.Is(err, domainkb.ErrInvalidID) {
		field = "id"
		text = "Données invalides : identifiant invalide (vide ou caractères interdits)."
	}
	return gomponents.Group([]gomponents.Node{
		errorMessage(text),
		gomponents.El("script", gomponents.Raw(focusFieldScript(field))),
	})
}

func focusFieldScript(fieldID string) string {
	return `(function(){var f=document.getElementById("field-` + fieldID + `");` +
		`if(!f){return;}f.classList.add("field-error");` +
		`f.addEventListener("input",function(){f.classList.remove("field-error");},{once:true});` +
		`f.focus();})();`
}

// duplicateMessage warns that a knowledge base already exists and links to edit.
func duplicateMessage(id string) gomponents.Node {
	return Span(Class("text-g-negative-text font-semibold"),
		gomponents.Text("La KB "+id+" existe déjà. "),
		A(Href("/kb/edit?id="+url.QueryEscape(id)),
			Class("text-g-accent-text underline"),
			gomponents.Text("Modifier cette KB ?"),
		),
	)
}

// redirectTo navigue vers target. POST natif : 303 (le navigateur suit).
// htmx : 200 + HX-Redirect — htmx suit une 303 et change la cible dans le
// DOM, le navigateur resterait donc sur l'URL d'origine (KB supprimée).
func redirectTo(w http.ResponseWriter, r *http.Request, target string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func redirectToDetail(w http.ResponseWriter, r *http.Request, id string) {
	redirectTo(w, r, "/kb/detail?id="+url.QueryEscape(id))
}

func (h *KBHandler) fetch(w http.ResponseWriter, r *http.Request, id string) (*domainkb.KB, bool) {
	item, err := h.app.Get(r.Context(), id)
	switch {
	case err == nil:
		return item, true
	case errors.Is(err, domainkb.ErrNotFound):
		writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
	default:
		writeInternalError(w, err)
	}
	return nil, false
}

// Detail displays a knowledge base and the related entries prepared by the application.
func (h *KBHandler) Detail(w http.ResponseWriter, r *http.Request) {
	detail, err := h.app.Detail(r.Context(), r.URL.Query().Get("id"))
	if err != nil {
		if errors.Is(err, domainkb.ErrNotFound) {
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", r.URL.Query().Get("id"))))
		} else {
			writeInternalError(w, err)
		}
		return
	}
	renderNode(w, pages.PageKBDetail(r.URL.Path, detail))
}

// Panel serves the drawer fragment (GET /kb/panel?id=…): it reuses the
// Detail use case but renders the standalone panel fragment instead of the
// full page — the drawer shell owns the page structure and scripts. A
// well-formed but unknown ID is 404, a malformed ID is 400 (no filesystem
// read ever happens).
func (h *KBHandler) Panel(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := domainkb.ValidateID(id); err != nil {
		writeMessage(w, http.StatusBadRequest, errorMessage(fmt.Sprintf("Identifiant KB invalide : %q", id)))
		return
	}

	detail, err := h.app.Detail(r.Context(), id)
	if err != nil {
		if errors.Is(err, domainkb.ErrNotFound) {
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
		} else {
			writeInternalError(w, err)
		}
		return
	}
	renderNode(w, pages.PageKBDetailPanel(detail))
}

func writeMessage(w http.ResponseWriter, status int, node gomponents.Node) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	renderNode(w, node)
}

func inputFromRequest(r *http.Request) dto.KBInput {
	return dto.KBInput{
		ID:         r.FormValue("id"),
		Erreur:     r.FormValue("erreur"),
		Notes:      r.FormValue("notes"),
		Resolution: r.FormValue("resolution"),
		Tags:       r.FormValue("tags"),
	}
}

func parseForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		writeMessage(w, http.StatusBadRequest, errorMessage("Formulaire invalide"))
		return false
	}
	return true
}

func renderNode(w http.ResponseWriter, node gomponents.Node) {
	if err := node.Render(w); err != nil {
		log.Printf("render HTML: %v", err)
	}
}

func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("application request failed: %v", err)
	writeMessage(w, http.StatusInternalServerError, errorMessage("Une erreur interne est survenue"))
}
