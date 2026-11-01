package kb

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MDRepository est l'adapter "fichier Markdown" du port kb.Repository :
// chaque KB est stockée dans le fichier data/kb/{ID}.md.
type MDRepository struct {
	dir string
}

// NewMDRepository crée un repository pointant sur le répertoire dir.
func NewMDRepository(dir string) *MDRepository {
	return &MDRepository{dir: dir}
}

func (r *MDRepository) path(id string) string {
	return filepath.Join(r.dir, id+".md")
}

// Save crée ou écrase le fichier {ID}.md (une mise à jour = réécriture).
func (r *MDRepository) Save(ctx context.Context, kb KB) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}

	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("kb md: création du répertoire: %w", err)
	}

	if err := os.WriteFile(r.path(kb.ID), []byte(kb.toMarkdown()), 0o644); err != nil {
		return fmt.Errorf("kb md: écriture de %s: %w", kb.ID, err)
	}

	return nil
}

// Exists retourne true si le fichier {ID}.md existe. L'ID est validé comme
// FindByID : il sert de nom de fichier.
func (r *MDRepository) Exists(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("kb md: %w", err)
	}

	if id == "" || !idPattern.MatchString(id) {
		return false, ErrInvalidID
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

// Delete supprime le fichier {ID}.md, ou retourne ErrNotFound si absente.
func (r *MDRepository) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("kb md: %w", err)
	}

	if id == "" || !idPattern.MatchString(id) {
		return ErrInvalidID
	}

	if err := os.Remove(r.path(id)); errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("kb md: %w: %s", ErrNotFound, id)
	} else if err != nil {
		return fmt.Errorf("kb md: suppression %s: %w", id, err)
	}

	return nil
}

// FindAll lit tous les fichiers .md du répertoire, triés par ID.
func (r *MDRepository) FindAll(ctx context.Context) ([]KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}

	entries, err := os.ReadDir(r.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []KB{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture du répertoire: %w", err)
	}

	kbs := []KB{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		id := strings.TrimSuffix(entry.Name(), ".md")
		kb, err := r.read(id)
		if err != nil {
			return nil, err
		}
		kbs = append(kbs, *kb)
	}

	sort.Slice(kbs, func(i, j int) bool { return kbs[i].ID < kbs[j].ID })

	return kbs, nil
}

// FindByID lit une KB par son identifiant.
func (r *MDRepository) FindByID(ctx context.Context, id string) (*KB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("kb md: %w", err)
	}

	// Même validation que le domaine : l'ID devient un nom de fichier.
	if id == "" || !idPattern.MatchString(id) {
		return nil, ErrInvalidID
	}

	return r.read(id)
}

func (r *MDRepository) read(id string) (*KB, error) {
	data, err := os.ReadFile(r.path(id))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("kb md: %w: %s", ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("kb md: lecture %s: %w", id, err)
	}

	kb := parseMarkdown(id, string(data))
	return &kb, nil
}
