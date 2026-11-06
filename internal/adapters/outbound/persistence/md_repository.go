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
	domainkb "grz/internal/domain/kb"
	"grz/internal/kbfile"
)

// MDRepository stores each knowledge base in a Markdown file.
type MDRepository struct {
	dir string
}

// NewMDRepository creates a repository rooted at dir.
func NewMDRepository(dir string) *MDRepository {
	return &MDRepository{dir: dir}
}

func (r *MDRepository) path(id string) string {
	return filepath.Join(r.dir, id+".md")
}

// Save creates or replaces the Markdown file for item.
func (r *MDRepository) Save(ctx context.Context, item domainkb.KB) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("kb md: création du répertoire: %w", err)
	}
	if err := os.WriteFile(r.path(item.ID), []byte(kbfile.Serialize(item)), 0o644); err != nil {
		return fmt.Errorf("kb md: écriture de %s: %w", item.ID, err)
	}
	return nil
}

// Exists reports whether the file for id exists.
func (r *MDRepository) Exists(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return false, domainkb.ErrInvalidID
	}
	_, err := os.Stat(r.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("kb md: stat %s: %w", id, err)
	}
	return true, nil
}

// Delete removes the file for id, or returns ErrNotFound when it is absent.
func (r *MDRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return domainkb.ErrInvalidID
	}
	if err := os.Remove(r.path(id)); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("kb md: %w: %s", domainkb.ErrNotFound, id)
	} else if err != nil {
		return fmt.Errorf("kb md: suppression %s: %w", id, err)
	}
	return nil
}

// FindAll reads Markdown files from the repository, sorted by ID.
func (r *MDRepository) FindAll(ctx context.Context) ([]domainkb.KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}
	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []domainkb.KB{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture du répertoire: %w", err)
	}

	items := make([]domainkb.KB, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		// Les fichiers du journal (journal-<date>.md) cohabitent dans le
		// même répertoire mais ne sont pas des KB : ni accueil ni graphe.
		if strings.HasPrefix(id, journalFilePrefix) {
			continue
		}
		item, err := r.read(id)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	// os.ReadDir returns entries sorted by filename; stripping the common .md
	// suffix preserves the same order by ID without a second sort.
	return items, nil
}

// FindByID reads one knowledge base by ID.
func (r *MDRepository) FindByID(ctx context.Context, id string) (*domainkb.KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}
	if err := domainkb.ValidateID(id); err != nil {
		return nil, domainkb.ErrInvalidID
	}
	return r.read(id)
}

func (r *MDRepository) read(id string) (*domainkb.KB, error) {
	data, err := os.ReadFile(r.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("kb md: %w: %s", domainkb.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture %s: %w", id, err)
	}
	item := kbfile.Parse(id, string(data))

	// Heure de dernière modification du fichier (mémoire de stockage, hors
	// format : rien n'est sérialisé dans le .md). Une stat qui échoue (ex:
	// fichier retiré entre ReadFile et Stat) laisse le champ à zéro — la
	// présentation sait masquer une date absente.
	if fi, statErr := os.Stat(r.path(id)); statErr == nil {
		item.Modified = fi.ModTime()
	}
	return &item, nil
}

var _ outbound.KBRepository = (*MDRepository)(nil)
