package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"grz/html/pages"
	"grz/internal/application"
	"grz/internal/application/dto"
	"grz/internal/application/ports/inbound"
	domainjournal "grz/internal/domain/journal"

	"maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// JournalHandler is the inbound HTTP adapter for the daily-journal
// use cases (Logseq-inspired: free markdown notes, one per day).
type JournalHandler struct {
	app inbound.JournalApplication
}

// NewJournalHandler injects the application port into the HTTP adapter.
func NewJournalHandler(app inbound.JournalApplication) *JournalHandler {
	return &JournalHandler{app: app}
}

// Page displays the journal: the note of the day (creatable in one click
// or openable when it exists), then the entries newest first.
func (h *JournalHandler) Page(w http.ResponseWriter, r *http.Request) {
	page, err := h.app.Page(r.Context(), journalToday())
	if err != nil {
		writeInternalError(w, err)
		return
	}
	renderNode(w, pages.PageJournal(r.URL.Path, page))
}

// journalToday fournit la date du calendrier local (l'application reste
// déterministe : la date vient de l'adapter, vive Logseq).
func journalToday() string {
	return time.Now().Format(domainjournal.DateLayout)
}

// ShowNewForm displays the journal-entry form, the date defaulting to
// today (a valid ?date= pre-fills an older day, to catch up on the journal).
func (h *JournalHandler) ShowNewForm(w http.ResponseWriter, r *http.Request) {
	valeurs := map[string]string{"date": journalToday()}
	if date := r.URL.Query().Get("date"); domainjournal.ValidateDate(date) == nil {
		valeurs["date"] = date
	}
	renderNode(w, pages.PageJournalCreateWith(r.URL.Path, pages.FormulaireJournal{Valeurs: valeurs}))
}

// Create validates and persists a new journal entry, rejecting to cover
// an already-written day (:409). After success, both flows return to the
// journal page — the Logseq flow: the writing is on the day-to-day board.
func (h *JournalHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	_, err := h.app.Create(r.Context(), journalInputFromRequest(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrAlreadyExists):
			date := strings.TrimSpace(r.FormValue("date"))
			message := journalDuplicateMessage(date)
			respondFormError(w, r, http.StatusConflict, message,
				pages.PageJournalCreateWith(r.URL.Path, pages.FormulaireJournal{Valeurs: journalFormValues(r), Message: message}))
		case errors.Is(err, domainjournal.ErrInvalidDate), errors.Is(err, domainjournal.ErrEmptyContent):
			message := journalValidationMessage(err)
			respondFormError(w, r, http.StatusBadRequest, message,
				pages.PageJournalCreateWith(r.URL.Path, pages.FormulaireJournal{Valeurs: journalFormValues(r), Message: message}))
		default:
			writeInternalError(w, err)
		}
		return
	}
	redirectTo(w, r, "/journal")
}

// ShowEntry displays a journal entry as a free note (markdown), with day
// navigation (previous/next) and edit/delete actions.
func (h *JournalHandler) ShowEntry(w http.ResponseWriter, r *http.Request) {
	page, err := h.app.Entry(r.Context(), r.PathValue("date"))
	if err != nil {
		switch {
		case errors.Is(err, domainjournal.ErrNotFound):
			date := r.PathValue("date")
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("Entrée de journal %q introuvable", date)))
		default:
			writeInternalError(w, err)
		}
		return
	}
	renderNode(w, pages.PageJournalEntry(r.URL.Path, page))
}

// ShowEditForm displays the pre-filled edit form: the frozen date (readonly
// field, the original value is validated and posted), the editable content.
func (h *JournalHandler) ShowEditForm(w http.ResponseWriter, r *http.Request) {
	page, err := h.app.Entry(r.Context(), r.PathValue("date"))
	if err != nil {
		switch {
		case errors.Is(err, domainjournal.ErrNotFound):
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("Entrée de journal %q introuvable", r.PathValue("date"))))
		default:
			writeInternalError(w, err)
		}
		return
	}
	renderNode(w, pages.PageJournalEdit(r.URL.Path, page.Entry))
}

// UpdateEntry validates and saves the edited content of an entry day.
func (h *JournalHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	entry, err := h.app.Update(r.Context(), journalInputFromRequest(r))
	if err != nil {
		if errors.Is(err, domainjournal.ErrInvalidDate) || errors.Is(err, domainjournal.ErrEmptyContent) {
			message := journalValidationMessage(err)
			respondFormError(w, r, http.StatusBadRequest, message,
				pages.PageJournalEditWith(r.URL.Path, strings.TrimSpace(r.FormValue("date")),
					pages.FormulaireJournal{Valeurs: journalFormValues(r), Message: message}))
		} else {
			writeInternalError(w, err)
		}
		return
	}
	redirectTo(w, r, "/journal/"+url.PathEscape(entry.Date))
}

// Delete removes an entry by date, then returns to the journal page.
func (h *JournalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	date := r.FormValue("date")
	if err := h.app.Delete(r.Context(), date); err != nil {
		switch {
		case errors.Is(err, domainjournal.ErrNotFound):
			writeMessage(w, http.StatusNotFound, errorMessage(fmt.Sprintf("Entrée de journal %q introuvable", date)))
		default:
			writeInternalError(w, err)
		}
		return
	}
	redirectTo(w, r, "/journal")
}

// journalFormValues collecte les valeurs saisies, indexées par name du
// champ, pour ré-afficher le formulaire sans perdre la saisie.
func journalFormValues(r *http.Request) map[string]string {
	valeurs := make(map[string]string, 2)
	valeurs["date"] = r.FormValue("date")
	valeurs["content"] = r.FormValue("content")
	return valeurs
}

// journalDuplicateMessage annonce qu'un jour est déjà écrit, avec un lien
// vers la note existante (entrée de la logique : jamais un état 409 muet).
func journalDuplicateMessage(date string) gomponents.Node {
	return Span(Class("text-g-negative-text font-semibold"),
		gomponents.Text("Une note existe déjà pour le "+date+". "),
		A(Href("/journal/"+url.QueryEscape(date)),
			Class("text-g-accent-text underline"),
			gomponents.Text("Ouvrir cette note ?"),
		),
	)
}

// journalValidationMessage identifie le champ fautif (date ou contenu) et
// demande son signalement thématique — aucune couleur inline.
func journalValidationMessage(err error) gomponents.Node {
	field := "content"
	text := "Données invalides : contenu de la note vide."
	if errors.Is(err, domainjournal.ErrInvalidDate) {
		field = "date"
		text = "Données invalides : date du journal invalide (format attendu AAAA-MM-JJ)."
	}
	return gomponents.Group([]gomponents.Node{
		errorMessage(text),
		gomponents.El("script", gomponents.Raw(focusFieldScript(field))),
	})
}

// journalInputFromRequest décode le formulaire (date + markdown libre).
func journalInputFromRequest(r *http.Request) dto.JournalInput {
	return dto.JournalInput{
		Date:    r.FormValue("date"),
		Content: r.FormValue("content"),
	}
}
