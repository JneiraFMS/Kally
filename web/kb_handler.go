package web

import (
	"errors"
	"fmt"
	"net/http"

	"grz/html/pages"
	"grz/internal/kb"
)

// KBHandler est l'adapter inbound HTTP du port kb.Repository :
// il traduit les requêtes en appels au repository, sans logique métier.
type KBHandler struct {
	repo kb.Repository
}

// NewKBHandler injecte le repository dans le handler.
func NewKBHandler(repo kb.Repository) *KBHandler {
	return &KBHandler{repo: repo}
}

// Home affiche la page d'accueil avec les cartes des KB existantes.
func (h *KBHandler) Home(w http.ResponseWriter, r *http.Request) {
	kbs, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.PageAccueil(r.URL.Path, kbs).Render(w)
}

// List affiche la page listant toutes les KB.
func (h *KBHandler) List(w http.ResponseWriter, r *http.Request) {
	kbs, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.PageKBList(r.URL.Path, kbs).Render(w)
}

// ShowNewForm affiche le formulaire de création.
func (h *KBHandler) ShowNewForm(w http.ResponseWriter, r *http.Request) {
	_ = pages.PageKBCreate(r.URL.Path).Render(w)
}

// Create enregistre une KB soumise par le formulaire (réponse HTMX partielle).
func (h *KBHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.saveForm(w, r, "")
}

// ShowEditForm affiche le formulaire de modification pré-rempli.
func (h *KBHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	kbFound, ok := h.fetch(w, r, r.URL.Query().Get("id"))
	if !ok {
		return
	}

	_ = pages.PageKBEdit(r.URL.Path, *kbFound).Render(w)
}

// UpdateForm enregistre les modifications d'une KB (réponse HTMX partielle).
func (h *KBHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	h.saveForm(w, r, "/kb/detail")
}

// saveForm factorise Create et UpdateForm.
// redirect est le lien "Consulter" affiché après succès (détail ou liste).
func (h *KBHandler) saveForm(w http.ResponseWriter, r *http.Request, detailNav string) {
	_ = r.ParseForm()

	newKB, err := kb.NewKB(
		r.FormValue("id"),
		r.FormValue("erreur"),
		r.FormValue("notes"),
		r.FormValue("resolution"),
	)
	if err != nil {
		writeMessage(w, http.StatusBadRequest,
			fmt.Sprintf("<span class='text-red-600 font-semibold'>Identifiant invalide : %s</span>", err))
		return
	}

	if err := h.repo.Save(r.Context(), newKB); err != nil {
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("<span class='text-red-600 font-semibold'>Erreur d'enregistrement : %s</span>", err))
		return
	}

	writeMessage(w, http.StatusOK,
		fmt.Sprintf(
			"<span class='text-green-600 font-semibold'>KB %s enregistrée ✅ <a class='text-blue-600 underline' href='%s?id=%s'>Consulter</a></span>",
			newKB.ID, detailNav, newKB.ID))
}

// fetch récupère une KB par son id de query, et écrit la réponse d'erreur
// sinon. Retourne false si la requête ne peut pas continuer.
func (h *KBHandler) fetch(w http.ResponseWriter, r *http.Request, id string) (*kb.KB, bool) {
	kbFound, err := h.repo.FindByID(r.Context(), id)

	switch {
	case err == nil:
		return kbFound, true
	case errors.Is(err, kb.ErrNotFound):
		writeMessage(w, http.StatusNotFound,
			fmt.Sprintf("<span class='text-red-600 font-semibold'>KB %q introuvable</span>", id))
		return nil, false
	default:
		writeMessage(w, http.StatusInternalServerError,
			fmt.Sprintf("<span class='text-red-600 font-semibold'>Erreur : %s</span>", err))
		return nil, false
	}
}

// Detail affiche la fiche complète d'une KB.
func (h *KBHandler) Detail(w http.ResponseWriter, r *http.Request) {
	kbFound, ok := h.fetch(w, r, r.URL.Query().Get("id"))
	if !ok {
		return
	}

	_ = pages.PageKBDetail(r.URL.Path, *kbFound).Render(w)
}

// writeMessage écrit une réponse HTML partielle (échange HTMX).
func writeMessage(w http.ResponseWriter, status int, html string) {
	w.WriteHeader(status)
	_, _ = fmt.Fprint(w, html)
}
