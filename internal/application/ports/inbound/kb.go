package inbound

import (
	"context"

	"grz/internal/application/dto"
	domainkb "grz/internal/domain/kb"
)

// KBApplication exposes the knowledge-base use cases to inbound adapters.
type KBApplication interface {
	List(ctx context.Context, tag string) ([]domainkb.KB, error)
	Get(ctx context.Context, id string) (*domainkb.KB, error)
	Detail(ctx context.Context, id string) (dto.KBDetail, error)
	Create(ctx context.Context, input dto.KBInput) (domainkb.KB, error)
	Update(ctx context.Context, input dto.KBInput) (domainkb.KB, error)
	Delete(ctx context.Context, id string) error
}
