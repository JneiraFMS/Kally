package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"grz/html/pages"
	"grz/internal/kb"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
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

// Home affiche la page d'accueil avec les cartes des KB existantes,
// filtrées par tag si ?tag=xxx est présent (page liste fusionnée ici).
func (h *KBHandler) Home(w http.ResponseWriter, r *http.Request) {
	kbs, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tag := r.URL.Query().Get("tag")
	if tag != "" {
		kbs = kb.FilterByTag(kbs, tag)
	}

	_ = pages.PageAccueil(r.URL.Path, kbs, tag).Render(w)
}

// List redirige vers l'accueil : la page liste a été fusionnée avec
// l'accueil (cartes + filtre ?tag=). Conservée pour les anciens liens/bookmarks.
func (h *KBHandler) List(w http.ResponseWriter, r *http.Request) {
	target := "/"
	if tag := r.URL.Query().Get("tag"); tag != "" {
		target = "/?tag=" + url.QueryEscape(tag)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// ShowGraph affiche la vue graphe globale des KB (SVG).
func (h *KBHandler) ShowGraph(w http.ResponseWriter, r *http.Request) {
	kbs, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = pages.PageKBGraph(r.URL.Path, kbs).Render(w)
}

// ShowNewForm affiche le formulaire de création.
func (h *KBHandler) ShowNewForm(w http.ResponseWriter, r *http.Request) {
	_ = pages.PageKBCreate(r.URL.Path).Render(w)
}

// Create enregistre une KB soumise par le formulaire de création.
// Un ID déjà présent est refusé (pas d'écrasement silencieux).
func (h *KBHandler) Create(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	newKB, err := kb.NewKB(
		r.FormValue("id"),
		r.FormValue("erreur"),
		r.FormValue("notes"),
		r.FormValue("resolution"),
		r.FormValue("tags"),
	)
	if err != nil {
		writeMessage(w, http.StatusBadRequest, validationMessage(err))
		return
	}

	exists, err := h.repo.Exists(r.Context(), newKB.ID)
	if err != nil {
		writeMessage(w, http.StatusInternalServerError, errorMessage("Erreur d'accès au stockage : "+err.Error()))
		return
	}
	if exists {
		writeMessage(w, http.StatusConflict, duplicateMessage(newKB.ID))
		return
	}

	if err := h.repo.Save(r.Context(), newKB); err != nil {
		writeMessage(w, http.StatusInternalServerError, errorMessage("Erreur d'enregistrement : "+err.Error()))
		return
	}

	redirectToDetail(w, r, newKB.ID)
}

// UpdateForm enregistre les modifications d'une KB (l'édition autorise
// réécrire le fichier existant, contrairement à la création).
func (h *KBHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	newKB, err := kb.NewKB(
		r.FormValue("id"),
		r.FormValue("erreur"),
		r.FormValue("notes"),
		r.FormValue("resolution"),
		r.FormValue("tags"),
	)
	if err != nil {
		writeMessage(w, http.StatusBadRequest, validationMessage(err))
		return
	}

	if err := h.repo.Save(r.Context(), newKB); err != nil {
		writeMessage(w, http.StatusInternalServerError, errorMessage("Erreur d'enregistrement : "+err.Error()))
		return
	}

	redirectToDetail(w, r, newKB.ID)
}

// ShowEditForm affiche le formulaire de modification pré-rempli.
func (h *KBHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	kbFound, ok := h.fetch(w, r, r.URL.Query().Get("id"))
	if !ok {
		return
	}

	_ = pages.PageKBEdit(r.URL.Path, *kbFound).Render(w)
}

// Delete supprime la KB id soumise en POST (formulaire natif, fonctionne
// sans JavaScript). En succès, ramène sur l'accueil.
func (h *KBHandler) Delete(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := r.FormValue("id")

	if err := h.repo.Delete(r.Context(), id); err != nil {
		switch {
		case errors.Is(err, kb.ErrNotFound):
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
		case errors.Is(err, kb.ErrInvalidID):
			writeMessage(w, http.StatusBadRequest, errorMessage("Identifiant de KB invalide"))
		default:
			writeMessage(w, http.StatusInternalServerError, errorMessage("Erreur de suppression : "+err.Error()))
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// errorMessage construit un message d'erreur neutre. gomponents échappe le
// texte : aucune donnée utilisateur (query ?id=) ne peut devenir du HTML.
func errorMessage(texte string) gomponents.Node {
	return Span(Class("text-g-negative font-semibold"), gomponents.Text(texte))
}

// validationMessage construit le message d'échec de validation du domaine et
// identifie le champ fautif pour l'entourer de rouge (script inline htmx).
func validationMessage(err error) gomponents.Node {
	field := "tags" // défaut : tag invalide
	texte := "Données invalides : " + err.Error()
	if errors.Is(err, kb.ErrInvalidID) {
		field = "id"
		texte = "Données invalides : identifiant invalide (vide ou caractères interdits)."
	}
	return gomponents.Group([]gomponents.Node{
		errorMessage(texte),
		gomponents.El("script", gomponents.Raw(focusFieldScript(field))),
	})
}

// focusFieldScript retourne un script mini : entoure de rouge le champ fautif
// et déplace le focus dessus. htmx exécute les scripts des contenus échangés
// (allowScriptTags par défaut) ; sans JS, dégradation douce (message seul).
func focusFieldScript(fieldID string) string {
	return `(function(){var f=document.getElementById("field-` + fieldID + `");` +
		`if(f){f.style.outline="2px solid #dc322f";f.style.outlineOffset="2px";f.focus();}})();`
}

// duplicateMessage avertit qu'une KB existe déjà et propose la modification.
func duplicateMessage(id string) gomponents.Node {
	return Span(Class("text-g-negative font-semibold"),
		gomponents.Text("La KB "+id+" existe déjà. "),
		A(Href("/kb/edit?id="+url.QueryEscape(id)),
			Class("text-g-accent underline"),
			gomponents.Text("Modifier cette KB ?"),
		),
	)
}

// redirectToDetail navigue vers la fiche de la KB enregistrée : htmx reçoit
// l'ordre via l'en-tête HX-Redirect, un formulaire natif suit la 303.
func redirectToDetail(w http.ResponseWriter, r *http.Request, id string) {
	target := "/kb/detail?id=" + url.QueryEscape(id)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", target)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// fetch récupère une KB par son id de query, et écrit la réponse d'erreur
// sinon. Retourne false si la requête ne peut pas continuer.
func (h *KBHandler) fetch(w http.ResponseWriter, r *http.Request, id string) (*kb.KB, bool) {
	kbFound, err := h.repo.FindByID(r.Context(), id)

	switch {
	case err == nil:
		return kbFound, true
	case errors.Is(err, kb.ErrNotFound):
		writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
		return nil, false
	default:
		writeMessage(w, http.StatusInternalServerError, errorMessage("Erreur : "+err.Error()))
		return nil, false
	}
}

// Detail affiche la fiche complète d'une KB avec ses KB liées par tags.
func (h *KBHandler) Detail(w http.ResponseWriter, r *http.Request) {
	// Le graphe est construit depuis l'ensemble des KB du repository.
	all, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id := r.URL.Query().Get("id")
	current := kb.ByID(all, id)
	if current == nil {
		writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("KB %q introuvable", id)))
		return
	}

	related := kb.Related(all, *current)
	_ = pages.PageKBDetail(r.URL.Path, *current, related).Render(w)
}

// writeMessage écrit une réponse HTML partielle (échange HTMX / page
// d'erreur). Le node est rendu par gomponents : le texte est échappé.
func writeMessage(w http.ResponseWriter, status int, node gomponents.Node) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = node.Render(w)
}
