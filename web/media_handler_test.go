package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"grz/internal/application/dto"
	"grz/internal/application/usecases"
	domainmedia "grz/internal/domain/media"
)

// fakeMediaRepo est un double en mémoire du port de stockage média.
type fakeMediaRepo struct {
	saved map[string][]byte
}

func newFakeMediaRepo() *fakeMediaRepo {
	return &fakeMediaRepo{saved: make(map[string][]byte)}
}

func (f *fakeMediaRepo) Save(_ context.Context, name string, content []byte) error {
	f.saved[name] = append([]byte(nil), content...)
	return nil
}

func (f *fakeMediaRepo) Find(_ context.Context, name string) ([]byte, error) {
	content, ok := f.saved[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domainmedia.ErrNotFound, name)
	}
	return content, nil
}

func newTestMediaHandler(repo *fakeMediaRepo) *MediaHandler {
	return NewMediaHandler(usecases.NewMediaService(repo))
}

// pngBytes commence par la signature PNG (http.DetectContentType → image/png).
var pngBytes = append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 64)...)

func TestMediaHandlerUploadStoresPNG(t *testing.T) {
	repo := newFakeMediaRepo()
	h := newTestMediaHandler(repo)

	req := httptest.NewRequest("POST", "/media", bytes.NewReader(pngBytes))
	req.Header.Set("Content-Type", "image/png")
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Upload() status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	var got dto.Media
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Upload() JSON invalide: %v", err)
	}
	if !strings.HasPrefix(got.URL, "/media/") {
		t.Errorf("Upload().URL = %q, want préfixe /media/", got.URL)
	}
	if !strings.HasSuffix(got.URL, ".png") {
		t.Errorf("Upload().URL = %q, want suffixe .png", got.URL)
	}
	if _, ok := repo.saved[got.Name]; !ok {
		t.Errorf("Upload() n'a pas stocké %q, saved = %v", got.Name, repo.saved)
	}
}

func TestMediaHandlerUploadRejectsNonImage(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	req := httptest.NewRequest("POST", "/media", strings.NewReader("pas une image"))
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Upload() status = %d, want 415", w.Code)
	}
}

func TestMediaHandlerUploadRejectsEmptyBody(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	req := httptest.NewRequest("POST", "/media", nil)
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Upload() status = %d, want 400", w.Code)
	}
}

func TestMediaHandlerUploadRejectsOversizedImage(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	big := make([]byte, domainmedia.MaxSize+1)
	copy(big, pngBytes)
	req := httptest.NewRequest("POST", "/media", bytes.NewReader(big))
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Upload() status = %d, want 413", w.Code)
	}
}

func TestMediaHandlerUploadRejectsBodyAboveHardLimit(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	huge := make([]byte, maxMediaBody+1)
	copy(huge, pngBytes)
	req := httptest.NewRequest("POST", "/media", bytes.NewReader(huge))
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Upload() status = %d, want 413", w.Code)
	}
}

func TestMediaHandlerServeStoredImage(t *testing.T) {
	repo := newFakeMediaRepo()
	h := newTestMediaHandler(repo)

	uploaded, err := usecases.NewMediaService(repo).Upload(context.Background(), pngBytes, "image/png")
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	req := httptest.NewRequest("GET", "/media/"+uploaded.Name, nil)
	req.SetPathValue("name", uploaded.Name)
	w := httptest.NewRecorder()

	h.Serve(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Serve() status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Serve() Content-Type = %q, want image/png", ct)
	}
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("Serve() X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := w.Header().Get("Cache-Control"); !strings.Contains(got, "immutable") {
		t.Errorf("Serve() Cache-Control = %q, want immutable", got)
	}
	if !bytes.Equal(w.Body.Bytes(), pngBytes) {
		t.Errorf("Serve() body = %q, want %q", w.Body.Bytes(), pngBytes)
	}
}

func TestMediaHandlerServeUnknownImage(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	req := httptest.NewRequest("GET", "/media/x", nil)
	req.SetPathValue("name", strings.Repeat("a", 64)+".png")
	w := httptest.NewRecorder()

	h.Serve(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Serve() status = %d, want 404", w.Code)
	}
}

func TestMediaHandlerServeRejectsInvalidName(t *testing.T) {
	h := newTestMediaHandler(newFakeMediaRepo())

	req := httptest.NewRequest("GET", "/media/../secret", nil)
	req.SetPathValue("name", "../secret.png")
	w := httptest.NewRecorder()

	h.Serve(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Serve() status = %d, want 400", w.Code)
	}
}

func TestMediaHandlerUploadPropagatesRepositoryError(t *testing.T) {
	repo := &failingMediaRepo{}
	h := NewMediaHandler(usecases.NewMediaService(repo))

	req := httptest.NewRequest("POST", "/media", bytes.NewReader(pngBytes))
	w := httptest.NewRecorder()

	h.Upload(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Upload() status = %d, want 500", w.Code)
	}
}

// failingMediaRepo simule une panne de stockage.
type failingMediaRepo struct{}

func (failingMediaRepo) Save(context.Context, string, []byte) error {
	return errors.New("disque plein")
}

func (failingMediaRepo) Find(context.Context, string) ([]byte, error) {
	return nil, errors.New("disque illisible")
}
