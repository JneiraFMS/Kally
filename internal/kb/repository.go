package kb

import "context"

// Repository est le port outbound du domaine KB : toute persistance
// (fichiers .md, base de données, ...) doit implémenter cette interface.
type Repository interface {
	// Save crée ou met à jour une KB.
	Save(ctx context.Context, kb KB) error

	// Exists retourne true si une KB avec cet identifiant existe déjà.
	Exists(ctx context.Context, id string) (bool, error)

	// FindAll retourne toutes les KB triées par ID.
	FindAll(ctx context.Context) ([]KB, error)

	// FindByID retourne une KB, ou ErrNotFound si absente.
	FindByID(ctx context.Context, id string) (*KB, error)

	// Delete supprime une KB, ou retourne ErrNotFound si absente.
	Delete(ctx context.Context, id string) error
}
