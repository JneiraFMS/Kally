package inbound

import (
	"context"

	"grz/internal/application/dto"
	domainkb "grz/internal/domain/kb"
)

// KBApplication exposes the knowledge-base use cases to inbound adapters.
type KBApplication interface {
	List(ctx context.Context, tags ...string) ([]domainkb.KB, error)
	Get(ctx context.Context, id string) (*domainkb.KB, error)
	Detail(ctx context.Context, id string) (dto.KBDetail, error)
	Create(ctx context.Context, input dto.KBInput) (domainkb.KB, error)
	Update(ctx context.Context, input dto.KBInput) (domainkb.KB, error)
	Delete(ctx context.Context, id string) error
	// Local library (File System Access): stateless transformations — the
	// server never persists locally; it serializes/analyzes and exposes its
	// shared IDs for rendering wiki links.
	SerializeLocal(input dto.KBInput) (string, error)
	ParseLocal(raw string) (dto.KBParsed, error)
	LocalKnownIDs(ctx context.Context) ([]string, error)
}
