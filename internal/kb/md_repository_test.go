package kb

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustKB(t *testing.T, id, erreur, notes, resolution string) KB {
	t.Helper()
	kb, err := NewKB(id, erreur, notes, resolution)
	if err != nil {
		t.Fatalf("NewKB(%q) inattendu: %v", id, err)
	}
	return kb
}

func TestMDRepositorySaveCreatesMarkdownFile(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	want := mustKB(t, "KB-0001", "Division par zéro", "Notes ici", "Résolution ici")

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "KB-0001.md"))
	if err != nil {
		t.Fatalf("fichier non créé: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "# KB-0001") || !strings.Contains(content, "Résolution ici") {
		t.Errorf("contenu .md invalide:\n%s", content)
	}
}

func TestMDRepositorySaveThenFindByID(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	want := mustKB(t, "KB-0009", "Erreur A", "Notes B", "Résolution C")
	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	got, err := repo.FindByID(ctx, "KB-0009")
	if err != nil {
		t.Fatalf("FindByID() inattendu: %v", err)
	}
	if got.ID != "KB-0009" || got.Erreur != "Erreur A" || got.Notes != "Notes B" || got.Resolution != "Résolution C" {
		t.Errorf("FindByID() = %+v, want %+v", *got, want)
	}
}

func TestMDRepositoryFindByIDNotFound(t *testing.T) {
	repo := NewMDRepository(t.TempDir())

	_, err := repo.FindByID(context.Background(), "KB-INCONNU")

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("FindByID() error = %v, want ErrNotFound", err)
	}
}

func TestMDRepositoryFindByIDRejectsInvalidID(t *testing.T) {
	repo := NewMDRepository(t.TempDir())

	_, err := repo.FindByID(context.Background(), "../secret")

	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("FindByID() error = %v, want ErrInvalidID", err)
	}
}

func TestMDRepositoryFindAllSorted(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	for _, id := range []string{"KB-0003", "KB-0001", "KB-0002"} {
		if err := repo.Save(ctx, mustKB(t, id, "e", "n", "r")); err != nil {
			t.Fatalf("Save(%q) inattendu: %v", id, err)
		}
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() inattendu: %v", err)
	}

	wantIDs := []string{"KB-0001", "KB-0002", "KB-0003"}
	for i, wantID := range wantIDs {
		if got[i].ID != wantID {
			t.Errorf("FindAll()[%d] = %q, want %q", i, got[i].ID, wantID)
		}
	}
}

func TestMDRepositoryFindAllEmptyDir(t *testing.T) {
	repo := NewMDRepository(filepath.Join(t.TempDir(), "inexistant"))

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("FindAll() inattendu: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("FindAll() = %d KB, want 0", len(got))
	}
}

func TestMDRepositoryFindAllIgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	if err := repo.Save(ctx, mustKB(t, "KB-0001", "e", "n", "r")); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.txt"), []byte("rien"), 0o644); err != nil {
		t.Fatalf("écriture fichier test: %v", err)
	}

	got, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() inattendu: %v", err)
	}
	if len(got) != 1 || got[0].ID != "KB-0001" {
		t.Errorf("FindAll() = %+v, want 1 KB KB-0001", got)
	}
}
