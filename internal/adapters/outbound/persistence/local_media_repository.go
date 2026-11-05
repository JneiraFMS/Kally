package persistence

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"grz/internal/application/ports/outbound"
	domainmedia "grz/internal/domain/media"
)

// LocalMediaRepository stores pasted images as files in a directory. Names
// are content-addressed and validated by the domain, so the adapter only
// joins them to its root directory.
type LocalMediaRepository struct {
	dir string
}

// NewLocalMediaRepository creates a media repository rooted at dir.
func NewLocalMediaRepository(dir string) *LocalMediaRepository {
	return &LocalMediaRepository{dir: dir}
}

func (r *LocalMediaRepository) path(name string) string {
	return filepath.Join(r.dir, name)
}

// Save writes content under name. The name being the hash of the content,
// an existing file is already correct and is left untouched.
func (r *LocalMediaRepository) Save(ctx context.Context, name string, content []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("media: %w", err)
	}
	if !domainmedia.ValidName(name) {
		return fmt.Errorf("media: %w: %s", domainmedia.ErrInvalidName, name)
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("media: création du répertoire: %w", err)
	}
	_, err := os.Stat(r.path(name))
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("media: stat %s: %w", name, err)
	}
	return writeFileAtomic(r.path(name), content)
}

// Find reads a stored image by name.
func (r *LocalMediaRepository) Find(ctx context.Context, name string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("media: %w", err)
	}
	if !domainmedia.ValidName(name) {
		return nil, fmt.Errorf("media: %w: %s", domainmedia.ErrInvalidName, name)
	}
	content, err := os.ReadFile(r.path(name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("media: %w: %s", domainmedia.ErrNotFound, name)
	}
	if err != nil {
		return nil, fmt.Errorf("media: lecture %s: %w", name, err)
	}
	return content, nil
}

// writeFileAtomic écrit dans un fichier temporaire du même répertoire puis le
// renomme : un média partiel n'est jamais visible sous son nom final.
func writeFileAtomic(path string, content []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".upload-*")
	if err != nil {
		return fmt.Errorf("media: fichier temporaire: %w", err)
	}
	tmpName := tmp.Name()

	// removeTemp supprime le fichier temporaire en conservant la cause
	// d'origine, et signale un éventuel échec de nettoyage.
	removeTemp := func(cause error) error {
		if removeErr := os.Remove(tmpName); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			return fmt.Errorf("%w (nettoyage: %v)", cause, removeErr)
		}
		return cause
	}

	if _, err := tmp.Write(content); err != nil {
		if closeErr := tmp.Close(); closeErr != nil {
			return removeTemp(fmt.Errorf("media: écriture: %w (fermeture: %v)", err, closeErr))
		}
		return removeTemp(fmt.Errorf("media: écriture: %w", err))
	}
	if err := tmp.Close(); err != nil {
		return removeTemp(fmt.Errorf("media: fermeture: %w", err))
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return removeTemp(fmt.Errorf("media: permissions: %w", err))
	}
	if err := os.Rename(tmpName, path); err != nil {
		return removeTemp(fmt.Errorf("media: renommage: %w", err))
	}
	return nil
}

var _ outbound.MediaRepository = (*LocalMediaRepository)(nil)
