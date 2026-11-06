package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"grz/html/pages"
	"grz/internal/application/ports/inbound"
	domainkb "grz/internal/domain/kb"
)

// maxLocalBody limite le corps des transformations texte de la
// bibliothèque locale (contenu .md lu par le script client).
const maxLocalBody = 1 << 20 // 1 Mio

// KBLocalHandler is the inbound HTTP adapter for the user's local library
// (File System Access). The server never persists anything there: it only
// contributes stateless transformations (serialize, parse, render), so a
// single codec remains the source of the file format.
type KBLocalHandler struct {
	app inbound.KBApplication
}

// NewKBLocalHandler injects the application port into the HTTP adapter.
func NewKBLocalHandler(app inbound.KBApplication) *KBLocalHandler {
	return &KBLocalHandler{app: app}
}

// Serialize answers POST /kb/local/serialize: the KB form fields, validated
// with the shared domain rules, become canonical .md content. The response
// body is the file the script writes on the user's machine.
func (h *KBLocalHandler) Serialize(w http.ResponseWriter, r *http.Request) {
	if !parseForm(w, r) {
		return
	}
	content, err := h.app.SerializeLocal(inputFromRequest(r))
	if err != nil {
		writePlainStatus(w, http.StatusBadRequest, "Enregistrement local refusé : "+plainValidationError(err))
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, content); err != nil {
		log.Printf("write local markdown: %v", err)
	}
}

// Parse answers POST /kb/local/parse: the .md content of a file read in the
// local library comes back broken down in JSON (fields of the detail card
// and pre-filled form), plus the server derived metadata (module, T-codes)
// computed with the exact same rules as the shared cards — the local card
// looks like any server card without duplicating the derivations
// client-side. The identity comes from the file name, not here.
func (h *KBLocalHandler) Parse(w http.ResponseWriter, r *http.Request) {
	raw, ok := readMarkdownBody(w, r)
	if !ok {
		return
	}
	parsed, err := h.app.ParseLocal(raw)
	if err != nil {
		log.Printf("parse local kb failed: %v", err)
		writePlainStatus(w, http.StatusInternalServerError, "Analyse impossible")
		return
	}
	type parsedDTO struct {
		Erreur     string   `json:"erreur"`
		Notes      string   `json:"notes"`
		Resolution string   `json:"resolution"`
		Tags       []string `json:"tags"`
		Module     string   `json:"module"`
		TCodes     []string `json:"tcodes"`
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(parsedDTO{
		Erreur:     parsed.Erreur,
		Notes:      parsed.Notes,
		Resolution: parsed.Resolution,
		Tags:       parsed.Tags,
		Module:     pages.ModuleSAPFor(parsed.Tags),
		TCodes:     pages.TCodesFor(parsed.Notes, parsed.Resolution),
	}); err != nil {
		log.Printf("encode parsed local kb: %v", err)
	}
}

// localRenderRequest is the body of POST /kb/local/render: the raw text
// sections of the file (the Erreur stays a raw trace, never rendered —
// the client displays it monospaced without the server pipeline).
type localRenderRequest struct {
	Notes      string `json:"notes"`
	Resolution string `json:"resolution"`
}

// localRenderResponse carries the rendered HTML fragments; both are empty
// strings when the section is empty (the fiche stays adaptive).
type localRenderResponse struct {
	NotesHTML      string `json:"notesHTML"`
	ResolutionHTML string `json:"resolutionHTML"`
}

// Render answers POST /kb/local/render: the markdown sections of a local
// file are rendered by the same goldmark pipeline as the server pages.
// Wiki links are resolved against the shared IDs: an existing target
// points to the record, an absent target becomes a dead link. Local files
// never enter the known set (watertight worlds).
func (h *KBLocalHandler) Render(w http.ResponseWriter, r *http.Request) {
	body := http.MaxBytesReader(w, r.Body, maxLocalBody)
	decoder := json.NewDecoder(body)
	var req localRenderRequest
	if err := decoder.Decode(&req); err != nil {
		writePlainStatus(w, http.StatusBadRequest, "Requête invalide")
		return
	}
	if strings.TrimSpace(req.Notes) == "" && strings.TrimSpace(req.Resolution) == "" {
		writePlainStatus(w, http.StatusBadRequest, "Contenu vide")
		return
	}
	known, err := h.app.LocalKnownIDs(r.Context())
	if err != nil {
		log.Printf("local known ids failed: %v", err)
		writePlainStatus(w, http.StatusInternalServerError, "Rendu impossible")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(localRenderResponse{
		NotesHTML:      pages.RenderMarkdownHTML(req.Notes, known),
		ResolutionHTML: pages.RenderMarkdownHTML(req.Resolution, known),
	}); err != nil {
		log.Printf("encode local render: %v", err)
	}
}

// readMarkdownBody lire le corps texte brut (maxLocalBody). Un corps vide
// ou surdimensionné est refusé en 400.
func readMarkdownBody(w http.ResponseWriter, r *http.Request) (string, bool) {
	body := http.MaxBytesReader(w, r.Body, maxLocalBody)
	data, err := io.ReadAll(body)
	if err != nil {
		writePlainStatus(w, http.StatusBadRequest, "Contenu impossible à lire")
		return "", false
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		writePlainStatus(w, http.StatusBadRequest, "Contenu vide")
		return "", false
	}
	return string(data), true
}

// writePlainStatus répond un message texte (consommé par le script client,
// jamais affiché tel quel).
func writePlainStatus(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	if _, err := io.WriteString(w, message); err != nil {
		log.Printf("write plain status: %v", err)
	}
}

// plainValidationError traduit une erreur de validation en message court,
// au même ton que les formulaires partagés.
func plainValidationError(err error) string {
	switch {
	case errors.Is(err, domainkb.ErrInvalidID):
		return "identifiant invalide (vide ou caractères interdits)."
	case errors.Is(err, domainkb.ErrInvalidTag):
		return "tag invalide (caractères interdits)."
	default:
		return fmt.Sprintf("données invalides (%v).", err)
	}
}
