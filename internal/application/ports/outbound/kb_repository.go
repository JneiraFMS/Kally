package outbound

import (
	"context"

	domainkb "grz/internal/domain/kb"
)

// KBRepository is the persistence port required by the KB application.
type KBRepository interface {
	Save(ctx context.Context, item domainkb.KB) error
	Exists(ctx context.Context, id string) (bool, error)
	FindAll(ctx context.Context) ([]domainkb.KB, error)
	FindByID(ctx context.Context, id string) (*domainkb.KB, error)
	Delete(ctx context.Context, id string) error
}
