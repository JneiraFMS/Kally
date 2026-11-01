package kb

import "context"

// Repository est le port outbound du domaine KB : toute persistance
// (fichiers .md, base de données, ...) doit implémenter cette interface.
type Repository interface {
	// Save crée ou met à jour une KB.
	Save(ctx context.Context, kb KB) error

	// FindAll retourne toutes les KB triées par ID.
	FindAll(ctx context.Context) ([]KB, error)

	// FindByID retourne une KB, ou ErrNotFound si absente.
	FindByID(ctx context.Context, id string) (*KB, error)
}
