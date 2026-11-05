package web

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"grz/internal/application/ports/inbound"
	domainmedia "grz/internal/domain/media"
)

// maxMediaBody borne le corps de la requête d'upload. La limite métier est
// domainmedia.MaxSize ; la marge couvre l'enveloppe HTTP avant le contrôle
// du domaine, qui reste l'autorité.
const maxMediaBody = domainmedia.MaxSize + 1<<20

// MediaHandler is the inbound HTTP adapter for pasted images: upload
// (POST /media) and serving (GET /media/{name}).
type MediaHandler struct {
	app inbound.MediaApplication
}

// NewMediaHandler injects the application port into the HTTP adapter.
func NewMediaHandler(app inbound.MediaApplication) *MediaHandler {
	return &MediaHandler{app: app}
}

// mediaError is the JSON error body returned to the paste script.
type mediaError struct {
	Error string `json:"error"`
}

// Upload stores an image sent as a raw body. The MIME type is recomputed
// from the binary signature: the client header is never trusted.
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMediaBody)
	content, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeMediaError(w, http.StatusRequestEntityTooLarge, "Image trop volumineuse (5 Mo maximum)")
			return
		}
		writeMediaError(w, http.StatusBadRequest, "Lecture de l'image impossible")
		return
	}
	if len(content) == 0 {
		writeMediaError(w, http.StatusBadRequest, "Aucune image reçue")
		return
	}

	item, err := h.app.Upload(r.Context(), content, http.DetectContentType(content))
	if err != nil {
		switch {
		case errors.Is(err, domainmedia.ErrUnsupportedType):
			writeMediaError(w, http.StatusUnsupportedMediaType, "Format non supporté (PNG, JPEG, GIF ou WebP)")
		case errors.Is(err, domainmedia.ErrEmpty):
			writeMediaError(w, http.StatusBadRequest, "Aucune image reçue")
		case errors.Is(err, domainmedia.ErrTooLarge):
			writeMediaError(w, http.StatusRequestEntityTooLarge, "Image trop volumineuse (5 Mo maximum)")
		default:
			log.Printf("media upload failed: %v", err)
			writeMediaError(w, http.StatusInternalServerError, "Une erreur interne est survenue")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(item); err != nil {
		log.Printf("media upload response: %v", err)
	}
}

// Serve returns a stored image. Names are content hashes, so the response
// can be cached forever; nosniff prevents any content-type guessing.
func (h *MediaHandler) Serve(w http.ResponseWriter, r *http.Request) {
	file, err := h.app.Get(r.Context(), r.PathValue("name"))
	if err != nil {
		switch {
		case errors.Is(err, domainmedia.ErrNotFound):
			http.NotFound(w, r)
		case errors.Is(err, domainmedia.ErrInvalidName):
			http.Error(w, "Nom d'image invalide", http.StatusBadRequest)
		default:
			log.Printf("media serve failed: %v", err)
			http.Error(w, "Une erreur interne est survenue", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", file.MIME)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	if _, err := w.Write(file.Content); err != nil {
		log.Printf("media serve response: %v", err)
	}
}

// writeMediaError writes a small JSON error body the paste script can show.
func writeMediaError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(mediaError{Error: message}); err != nil {
		log.Printf("media error response: %v", err)
	}
}
