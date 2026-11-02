package persistence

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	domainkb "grz/internal/domain/kb"
)

func mustKB(t *testing.T, id, erreur, notes, resolution string, tags ...string) domainkb.KB {
	t.Helper()
	kb, err := domainkb.NewKB(id, erreur, notes, resolution, tags...)
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

	if !errors.Is(err, domainkb.ErrNotFound) {
		t.Errorf("FindByID() error = %v, want ErrNotFound", err)
	}
}

func TestMDRepositoryFindByIDRejectsInvalidID(t *testing.T) {
	repo := NewMDRepository(t.TempDir())

	_, err := repo.FindByID(context.Background(), "../secret")

	if !errors.Is(err, domainkb.ErrInvalidID) {
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

func TestMDRepositorySaveWritesFrontMatter(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)

	want := mustKB(t, "KB-0001", "e", "n", "r", "incident", "kpi")
	if err := repo.Save(context.Background(), want); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "KB-0001.md"))
	if err != nil {
		t.Fatalf("lecture: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "---\ntags: [incident, kpi]\n---\n") {
		t.Errorf("front matter manquant en tête de fichier:\n%s", content)
	}
}

func TestMDRepositoryRoundTripTags(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	want := mustKB(t, "KB-0002", "e", "n", "r", "incident", "calcul")
	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	got, err := repo.FindByID(ctx, "KB-0002")
	if err != nil {
		t.Fatalf("FindByID() inattendu: %v", err)
	}
	if !slices.Equal(got.Tags, []string{"incident", "calcul"}) {
		t.Errorf("FindByID() Tags = %v, want [incident calcul]", got.Tags)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() inattendu: %v", err)
	}
	if !slices.Equal(all[0].Tags, []string{"incident", "calcul"}) {
		t.Errorf("FindAll() Tags = %v, want [incident calcul]", all[0].Tags)
	}
}

func TestMDRepositoryBackwardCompatibleWithoutFrontMatter(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)

	// Fichier ancien format (sans front matter) écrit à la main
	content := "---\nnote: autre chose\n---\n# KB-OLD\n\n## Erreur\nancienne\n\n## Notes\nn\n\n## Résolution\nr\n"
	// NB: "note:" dans le front matter doit être ignoré (pas de tags)
	if err := os.WriteFile(filepath.Join(dir, "KB-OLD.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("écriture: %v", err)
	}

	got, err := repo.FindByID(context.Background(), "KB-OLD")
	if err != nil {
		t.Fatalf("FindByID() inattendu: %v", err)
	}
	if len(got.Tags) != 0 {
		t.Errorf("FindByID() Tags = %v, want vide (pas de front matter tags)", got.Tags)
	}
	if got.Erreur != "ancienne" {
		t.Errorf("FindByID() Erreur = %q, want \"ancienne\"", got.Erreur)
	}
}

// TestMDRepositoryExists : Exists distingue fichier présent / absent / ID
// invalide (l'ID sert de nom de fichier, donc normalization identique).
func TestMDRepositoryExists(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, "KB-0001")
	if err != nil {
		t.Fatalf("Exists() inattendu: %v", err)
	}
	if exists {
		t.Error("Exists() = true sur répertoire vide, want false")
	}

	if err := repo.Save(ctx, mustKB(t, "KB-0001", "e", "n", "r")); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	exists, err = repo.Exists(ctx, "KB-0001")
	if err != nil {
		t.Fatalf("Exists() inattendu: %v", err)
	}
	if !exists {
		t.Error("Exists() = false après Save(), want true")
	}

	exists, err = repo.Exists(ctx, "../secret")
	if !errors.Is(err, domainkb.ErrInvalidID) {
		t.Errorf("Exists() error = %v, want ErrInvalidID", err)
	}
	if exists {
		t.Error("Exists() = true pour un ID invalide, want false")
	}
}

// TestMDRepositoryDeleteRemovesFile : Delete supprime le .md puis FindByID
// doit tomber en ErrNotFound.
func TestMDRepositoryDeleteRemovesFile(t *testing.T) {
	dir := t.TempDir()
	repo := NewMDRepository(dir)
	ctx := context.Background()

	if err := repo.Save(ctx, mustKB(t, "KB-0001", "e", "n", "r")); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	if err := repo.Delete(ctx, "KB-0001"); err != nil {
		t.Fatalf("Delete() inattendu: %v", err)
	}

	if _, err := repo.FindByID(ctx, "KB-0001"); !errors.Is(err, domainkb.ErrNotFound) {
		t.Errorf("FindByID() après Delete error = %v, want ErrNotFound", err)
	}

	// Le répertoire ne doit plus contenir que ce qu'on y a mis d'autre
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Delete() a laissé des fichiers: %v", entries)
	}
}

// TestMDRepositoryDeleteNotFound : supprimer une KB absente est une erreur
// ErrNotFound (pas un succès silencieux).
func TestMDRepositoryDeleteNotFound(t *testing.T) {
	repo := NewMDRepository(t.TempDir())

	if err := repo.Delete(context.Background(), "KB-INCONNU"); !errors.Is(err, domainkb.ErrNotFound) {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}

// TestMDRepositoryDeleteRejectsInvalidID : l'ID sert de nom de fichier,
// Delete applique la même validation que FindByID.
func TestMDRepositoryDeleteRejectsInvalidID(t *testing.T) {
	repo := NewMDRepository(t.TempDir())

	if err := repo.Delete(context.Background(), "../secret"); !errors.Is(err, domainkb.ErrInvalidID) {
		t.Errorf("Delete() error = %v, want ErrInvalidID", err)
	}
}

func BenchmarkMDRepositoryFindAll(b *testing.B) {
	repo := NewMDRepository(b.TempDir())
	for i := 0; i < 500; i++ {
		item, err := domainkb.NewKB(fmt.Sprintf("KB-%04d", i), "error", "notes", "resolution", "incident", "kpi")
		if err != nil {
			b.Fatal(err)
		}
		if err := repo.Save(context.Background(), item); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.FindAll(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}
