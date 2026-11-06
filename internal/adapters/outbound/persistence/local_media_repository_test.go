package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainmedia "grz/internal/domain/media"
)

// mediaName retourne un nom valide déterministe pour les tests.
func mediaName(seed string) string {
	return strings.Repeat(seed, 64/len(seed)) + ".png"
}

func TestLocalMediaRepositorySaveThenFind(t *testing.T) {
	repo := NewLocalMediaRepository(t.TempDir())
	ctx := context.Background()
	name := mediaName("ab")
	content := []byte{0x89, 'P', 'N', 'G'}

	if err := repo.Save(ctx, name, content); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	got, err := repo.Find(ctx, name)
	if err != nil {
		t.Fatalf("Find(): %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Find() = %q, want %q", got, content)
	}
}

func TestLocalMediaRepositorySaveCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "media")
	repo := NewLocalMediaRepository(dir)

	if err := repo.Save(context.Background(), mediaName("cd"), []byte("png")); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, mediaName("cd"))); err != nil {
		t.Errorf("fichier média absent: %v", err)
	}
}

func TestLocalMediaRepositorySaveIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	repo := NewLocalMediaRepository(dir)
	ctx := context.Background()
	name := mediaName("ef")
	content := []byte("png")

	if err := repo.Save(ctx, name, content); err != nil {
		t.Fatalf("Save(): %v", err)
	}
	if err := repo.Save(ctx, name, content); err != nil {
		t.Fatalf("Save() second appel: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(): %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("fichiers = %d, want 1 (déduplication)", len(entries))
	}
}

func TestLocalMediaRepositoryRejectsInvalidName(t *testing.T) {
	repo := NewLocalMediaRepository(t.TempDir())
	ctx := context.Background()

	if err := repo.Save(ctx, "../secret.png", []byte("x")); !errors.Is(err, domainmedia.ErrInvalidName) {
		t.Errorf("Save(invalide) error = %v, want ErrInvalidName", err)
	}
	if _, err := repo.Find(ctx, "../secret.png"); !errors.Is(err, domainmedia.ErrInvalidName) {
		t.Errorf("Find(invalide) error = %v, want ErrInvalidName", err)
	}
}

func TestLocalMediaRepositoryFindUnknown(t *testing.T) {
	repo := NewLocalMediaRepository(t.TempDir())

	if _, err := repo.Find(context.Background(), mediaName("01")); !errors.Is(err, domainmedia.ErrNotFound) {
		t.Errorf("Find(inconnu) error = %v, want ErrNotFound", err)
	}
}
