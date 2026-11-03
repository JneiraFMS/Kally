package persistence

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainjournal "grz/internal/domain/journal"
)

func newTestJournalDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// TestMDJournalSaveAndByDate : une entrée est persistée dans un fichier
// journal-<date>.md dont le contenu EST le markdown de la note (pas de
// surcharge KB : pas de sections Erreur/Notes/Résolution).
func TestMDJournalSaveAndByDate(t *testing.T) {
	dir := newTestJournalDir(t)
	repo := NewMDJournalRepository(dir)
	ctx := context.Background()

	entry, err := domainjournal.New("2026-11-03", "**Une** note libre\n\n- item 1")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if err := repo.Save(ctx, entry); err != nil {
		t.Fatalf("Save() inattendu: %v", err)
	}

	// Le fichier porte le contenu brut, sans entête
	raw, err := os.ReadFile(filepath.Join(dir, "journal-2026-11-03.md"))
	if err != nil {
		t.Fatalf("fichier journal manquant: %v", err)
	}
	if !strings.Contains(string(raw), "**Une** note libre") {
		t.Errorf("le markdown libre n'est pas stocké tel quel: %q", raw)
	}
	if strings.Contains(string(raw), "## Erreur") {
		t.Errorf("format KB (sections Erreur/Notes/Résolution) leaké dans le journal: %q", raw)
	}

	loaded, err := repo.ByDate(ctx, "2026-11-03")
	if err != nil {
		t.Fatalf("ByDate() inattendu: %v", err)
	}
	if loaded.Date != "2026-11-03" || loaded.Content != entry.Content {
		t.Errorf("ByDate() = %+v, want contenu tel quel", loaded)
	}

	if _, err := repo.ByDate(ctx, "2026-10-30"); !errors.Is(err, domainjournal.ErrNotFound) {
		t.Errorf("ByDate() absent = %v, want ErrNotFound", err)
	}
}

// TestMDJournalExistsAndEntries : Exists distingue présent/absent, Entries
// liste toutes les entrées de la plus récente à la plus ancienne, en
// ignorant les fichiers qui ne sont pas des entrées de journal.
func TestMDJournalExistsAndEntries(t *testing.T) {
	dir := newTestJournalDir(t)
	repo := NewMDJournalRepository(dir)
	ctx := context.Background()

	if ok, err := repo.Exists(ctx, "2026-11-03"); err != nil || ok {
		t.Fatalf("Exists() avant écriture = (%v, %v), want (false, nil)", ok, err)
	}

	_ = mustSaved(t, repo, ctx, "2026-11-03", "recente")
	_ = mustSaved(t, repo, ctx, "2026-01-15", "ancienne")

	if ok, err := repo.Exists(ctx, "2026-11-03"); err != nil || !ok {
		t.Errorf("Exists() = (%v, %v), want (true, nil)", ok, err)
	}

	entries, err := repo.Entries(ctx)
	if err != nil {
		t.Fatalf("Entries() inattendu: %v", err)
	}
	if len(entries) != 2 || entries[0].Date != "2026-11-03" || entries[1].Date != "2026-01-15" {
		t.Errorf("Entries() = %+v, want 2026-11-03 puis 2026-01-15", entries)
	}
}

// TestMDJournalIgnoresForeignFiles : le répertoire est partagé avec les KB
// — un fichier KB.md ou un journal- au nom invalide ne doit pas faire
// échouer la liste du journal.
func TestMDJournalIgnoresForeignFiles(t *testing.T) {
	dir := newTestJournalDir(t)
	repo := NewMDJournalRepository(dir)
	ctx := context.Background()

	_ = mustSaved(t, repo, ctx, "2026-11-03", "une seule entrée")

	// Une KB (sans préfixe journal) et un fichier au nom corrompu cohabitent
	foreign := "# KB-0001\n\n## Erreur\nx\n"
	if err := os.WriteFile(filepath.Join(dir, "KB-0001.md"), []byte(foreign), 0o644); err != nil {
		t.Fatalf("écriture KB étrangère: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "journal-corrompu.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("écriture corrompue: %v", err)
	}

	entries, err := repo.Entries(ctx)
	if err != nil {
		t.Fatalf("Entries() a échoué sur des fichiers étrangers: %v", err)
	}
	if len(entries) != 1 || entries[0].Date != "2026-11-03" {
		t.Errorf("Entries() = %+v, want seulement 2026-11-03", entries)
	}
}

// TestMDJournalDelete : supprimer une entrée existante retire son fichier ;
// une entrée absente remonte ErrNotFound.
func TestMDJournalDelete(t *testing.T) {
	dir := newTestJournalDir(t)
	repo := NewMDJournalRepository(dir)
	ctx := context.Background()

	_ = mustSaved(t, repo, ctx, "2026-11-03", "x")

	if err := repo.Delete(ctx, "2026-11-03"); err != nil {
		t.Fatalf("Delete() inattendu: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "journal-2026-11-03.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Delete() a laissé le fichier")
	}

	if err := repo.Delete(ctx, "2026-11-03"); !errors.Is(err, domainjournal.ErrNotFound) {
		t.Errorf("Delete() absent = %v, want ErrNotFound", err)
	}
}

// TestMDJournalDateInsecurisee : une date invalide ne doit jamais atteindre
// le disque (directory traversal interdit : le nom de fichier vient de la
// date).
func TestMDJournalDateInsecurisee(t *testing.T) {
	dir := newTestJournalDir(t)
	repo := NewMDJournalRepository(dir)
	ctx := context.Background()

	if ok, err := repo.Exists(ctx, "2026-01-01/.."); ok || err == nil {
		t.Errorf("Exists() date frauduleuse = (%v, %v), want refus", ok, err)
	}
	if _, err := repo.Exists(ctx, ""); err == nil {
		t.Errorf("Exists() date vide = pas d'erreur, want ErrInvalidDate")
	}
}

// TestMDKBRepositorySkipsJournalFiles : dans le répertoire partagé, la
// liste des KB ignore les fichiers journal (ce ne sont pas des KB : ni
// accueil ni graphe.)
func TestMDKBRepositorySkipsJournalFiles(t *testing.T) {
	dir := newTestJournalDir(t)
	kbRepo := NewMDRepository(dir)
	journalRepo := NewMDJournalRepository(dir)
	ctx := context.Background()

	journalEntry, err := domainjournal.New("2026-11-03", "note libre")
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if err := journalRepo.Save(ctx, journalEntry); err != nil {
		t.Fatalf("Save() journal: %v", err)
	}

	items, err := kbRepo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() a échoué avec un fichier journal présent: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("FindAll() inclut les fichiers journal: %+v", items)
	}
}

func mustSaved(t *testing.T, repo *MDJournalRepository, ctx context.Context, date, content string) domainjournal.Entry {
	t.Helper()
	entry, err := domainjournal.New(date, content)
	if err != nil {
		t.Fatalf("New(%q): %v", date, err)
	}
	if err := repo.Save(ctx, entry); err != nil {
		t.Fatalf("Save(%q): %v", date, err)
	}
	return entry
}
