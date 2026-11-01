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

// Home displays the knowledge-base cards, optionally filtered by tag.
func (h *KBHandler) Home(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	items, err := h.app.List(r.Context(), tag)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	renderNode(w, pages.PageAccueil(r.URL.Path, items, tag))
}

// List redirects old list URLs to the home page.
func (h *KBHandler) List(w http.ResponseWriter, r *http.Request) {
	target := "/"
	if tag := r.URL.Query().Get("tag"); tag != "" {
		target = "/?tag=" + url.QueryEscape(tag)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// ShowGraph displays the global knowledge-base graph.
func (h *KBHandler) ShowGraph(w http.ResponseWriter, r *http.Request) {
	items, err := h.app.List(r.Context(), "")
	if err != nil {
		writeInternalError(w, err)
		return
	}
	renderNode(w, pages.PageKBGraph(r.URL.Path, items))
}

// ShowNewForm displays the create form.
func (h *KBHandler) ShowNewForm(w http.ResponseWriter, r *http.Request) {
	renderNode(w, pages.PageKBCreate(r.URL.Path))
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
			writeMessage(w, http.StatusConflict, duplicateMessage(strings.TrimSpace(r.FormValue("id"))))
		case errors.Is(err, domainkb.ErrInvalidID), errors.Is(err, domainkb.ErrInvalidTag):
			writeMessage(w, http.StatusBadRequest, validationMessage(err))
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
			writeMessage(w, http.StatusBadRequest, validationMessage(err))
		} else {
			writeInternalError(w, err)
		}
		return
	}
	redirectToDetail(w, r, item.ID)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// errorMessage builds a neutral HTML message; gomponents escapes its text.
func errorMessage(text string) gomponents.Node {
	return Span(Class("text-g-negative font-semibold"), gomponents.Text(text))
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
		`if(f){f.style.outline="2px solid #dc322f";f.style.outlineOffset="2px";f.focus();}})();`
}

// duplicateMessage warns that a knowledge base already exists and links to edit.
func duplicateMessage(id string) gomponents.Node {
	return Span(Class("text-g-negative font-semibold"),
		gomponents.Text("La KB "+id+" existe déjà. "),
		A(Href("/kb/edit?id="+url.QueryEscape(id)),
			Class("text-g-accent underline"),
			gomponents.Text("Modifier cette KB ?"),
		),
	)
}

func redirectToDetail(w http.ResponseWriter, r *http.Request, id string) {
	target := "/kb/detail?id=" + url.QueryEscape(id)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
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
	renderNode(w, pages.PageKBDetail(r.URL.Path, detail.KB, detail.Related))
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
