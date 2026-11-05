package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	domainmedia "grz/internal/domain/media"
)

// fakeMediaRepository est un double en mémoire du port de stockage média.
type fakeMediaRepository struct {
	saved     map[string][]byte
	saveErr   error
	findErr   error
	findCalls int
}

func newFakeMediaRepository() *fakeMediaRepository {
	return &fakeMediaRepository{saved: make(map[string][]byte)}
}

func (r *fakeMediaRepository) Save(_ context.Context, name string, content []byte) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.saved[name] = append([]byte(nil), content...)
	return nil
}

func (r *fakeMediaRepository) Find(_ context.Context, name string) ([]byte, error) {
	r.findCalls++
	if r.findErr != nil {
		return nil, r.findErr
	}
	content, ok := r.saved[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", domainmedia.ErrNotFound, name)
	}
	return content, nil
}

var testPNG = append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, []byte("payload")...)

func TestMediaServiceUploadStoresContentAddressedImage(t *testing.T) {
	repo := newFakeMediaRepository()
	service := NewMediaService(repo)

	got, err := service.Upload(context.Background(), testPNG, "image/png")
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	name, err := domainmedia.NameFor("image/png", testPNG)
	if err != nil {
		t.Fatalf("NameFor(): %v", err)
	}
	if got.Name != name {
		t.Errorf("Upload().Name = %q, want %q", got.Name, name)
	}
	if got.URL != "/media/"+name {
		t.Errorf("Upload().URL = %q, want %q", got.URL, "/media/"+name)
	}
	if stored, ok := repo.saved[name]; !ok {
		t.Errorf("Upload() n'a pas stocké %q, saved = %v", name, repo.saved)
	} else if string(stored) != string(testPNG) {
		t.Errorf("contenu stocké = %q, want %q", stored, testPNG)
	}
}

func TestMediaServiceUploadRejectsUnsupportedType(t *testing.T) {
	repo := newFakeMediaRepository()
	service := NewMediaService(repo)

	_, err := service.Upload(context.Background(), []byte("<svg/>"), "image/svg+xml")
	if !errors.Is(err, domainmedia.ErrUnsupportedType) {
		t.Errorf("Upload() error = %v, want ErrUnsupportedType", err)
	}
	if len(repo.saved) != 0 {
		t.Errorf("Upload() a stocké malgré le refus: %v", repo.saved)
	}
}

func TestMediaServiceUploadPropagatesRepositoryError(t *testing.T) {
	repo := newFakeMediaRepository()
	repo.saveErr = errors.New("disque plein")
	service := NewMediaService(repo)

	if _, err := service.Upload(context.Background(), testPNG, "image/png"); err == nil {
		t.Error("Upload() error = nil, want l'erreur du repository")
	}
}

func TestMediaServiceGetReturnsContentAndMIME(t *testing.T) {
	repo := newFakeMediaRepository()
	service := NewMediaService(repo)
	uploaded, err := service.Upload(context.Background(), testPNG, "image/png")
	if err != nil {
		t.Fatalf("Upload(): %v", err)
	}

	file, err := service.Get(context.Background(), uploaded.Name)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}
	if file.MIME != "image/png" {
		t.Errorf("Get().MIME = %q, want image/png", file.MIME)
	}
	if string(file.Content) != string(testPNG) {
		t.Errorf("Get().Content = %q, want %q", file.Content, testPNG)
	}
}

func TestMediaServiceGetRejectsInvalidNameWithoutRepository(t *testing.T) {
	repo := newFakeMediaRepository()
	service := NewMediaService(repo)

	_, err := service.Get(context.Background(), "../secret.png")
	if !errors.Is(err, domainmedia.ErrInvalidName) {
		t.Errorf("Get() error = %v, want ErrInvalidName", err)
	}
	if repo.findCalls != 0 {
		t.Errorf("Get() a interrogé le repository %d fois, want 0", repo.findCalls)
	}
}

func TestMediaServiceGetPropagatesNotFound(t *testing.T) {
	repo := newFakeMediaRepository()
	service := NewMediaService(repo)

	_, err := service.Get(context.Background(), strings.Repeat("a", 64)+".png")
	if !errors.Is(err, domainmedia.ErrNotFound) {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}
