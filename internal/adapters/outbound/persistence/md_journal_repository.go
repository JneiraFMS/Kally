package persistence

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"grz/internal/application/ports/outbound"
	domainjournal "grz/internal/domain/journal"
)

// journalFilePrefix est la convention de nommage des fichiers du journal
// dans le répertoire partagé avec les KB : une entrée = journal-<date>.md.
const journalFilePrefix = "journal-"

// MDJournalRepository stores one journal entry per Markdown file, named
// journal-<date>.md. The file content is the free markdown of the note —
// the date lives in the filename, no KB face (sections/tags).
type MDJournalRepository struct {
	dir string
}

// NewMDJournalRepository creates a journal repository rooted at dir (the
// directory shared with the KB repository).
func NewMDJournalRepository(dir string) *MDJournalRepository {
	return &MDJournalRepository{dir: dir}
}

func (r *MDJournalRepository) path(date string) string {
	return filepath.Join(r.dir, journalFilePrefix+date+".md")
}

// Save creates or replaces the file of the entry's day.
func (r *MDJournalRepository) Save(ctx context.Context, entry domainjournal.Entry) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("journal md: %w", err)
	}
	// date validée par le domaine à l'entrée ; contrôle défensif ici :
	// la date devient un nom de fichier, refuse celle qui ne l'est pas.
	if err := domainjournal.ValidateDate(entry.Date); err != nil {
		return err
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("journal md: création du répertoire: %w", err)
	}
	if err := os.WriteFile(r.path(entry.Date), []byte(entry.Content), 0o644); err != nil {
		return fmt.Errorf("journal md: écriture de %s: %w", entry.Date, err)
	}
	return nil
}

// Exists reports whether the entry's file exists.
func (r *MDJournalRepository) Exists(ctx context.Context, date string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("journal md: %w", err)
	}
	if err := domainjournal.ValidateDate(date); err != nil {
		return false, err
	}
	_, err := os.Stat(r.path(date))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("journal md: stat %s: %w", date, err)
	}
	return true, nil
}

// ByDate reads one entry by date, ErrNotFound when absent.
func (r *MDJournalRepository) ByDate(ctx context.Context, date string) (*domainjournal.Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("journal md: %w", err)
	}
	if err := domainjournal.ValidateDate(date); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(r.path(date))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", domainjournal.ErrNotFound, date)
	}
	if err != nil {
		return nil, fmt.Errorf("journal md: lecture %s: %w", date, err)
	}
	if err := requireContent(date, string(data)); err != nil {
		return nil, err
	}
	return &domainjournal.Entry{Date: date, Content: string(data)}, nil
}

// Entries reads all journal entries, ignoring files that are not entries
// (KB files, or a corrupted journal name recorded by hand).
func (r *MDJournalRepository) Entries(ctx context.Context) ([]domainjournal.Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("journal md: %w", err)
	}
	dirEntries, err := os.ReadDir(r.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []domainjournal.Entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("journal md: lecture du répertoire: %w", err)
	}

	entries := make([]domainjournal.Entry, 0, len(dirEntries))
	for _, d := range dirEntries {
		if d.IsDir() || !strings.HasPrefix(d.Name(), journalFilePrefix) || !strings.HasSuffix(d.Name(), ".md") {
			continue
		}
		date := strings.TrimSuffix(strings.TrimPrefix(d.Name(), journalFilePrefix), ".md")
		entry, err := r.ByDate(ctx, date)
		if err != nil {
			// nom de fichier invalide ou fichier vide (édité à la main) :
			// ignoré, comme un fichier étranger (le journal reste sûr par
			// défaut).
			if errors.Is(err, domainjournal.ErrInvalidDate) || errors.Is(err, domainjournal.ErrEmptyContent) {
				continue
			}
			return nil, err
		}
		entries = append(entries, *entry)
	}
	return domainjournal.EntriesDesc(entries), nil
}

// Delete removes the entry's file, or returns ErrNotFound when absent.
func (r *MDJournalRepository) Delete(ctx context.Context, date string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("journal md: %w", err)
	}
	if err := domainjournal.ValidateDate(date); err != nil {
		return err
	}
	if err := os.Remove(r.path(date)); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("journal md: %w: %s", domainjournal.ErrNotFound, date)
	} else if err != nil {
		return fmt.Errorf("journal md: suppression %s: %w", date, err)
	}
	return nil
}

// requireContent rejette un fichier d'entrée vide manuellement corrompu.
func requireContent(date, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("%w: %s", domainjournal.ErrEmptyContent, date)
	}
	return nil
}

var _ outbound.JournalRepository = (*MDJournalRepository)(nil)
