package inbound

import (
	"context"

	"grz/internal/application/dto"
)

// MediaApplication exposes the pasted-image use cases to inbound adapters.
// The MIME type is detected by the adapter (binary signature) before being
// handed to the application, which owns the validation rules.
type MediaApplication interface {
	Upload(ctx context.Context, content []byte, mime string) (dto.Media, error)
	Get(ctx context.Context, name string) (dto.MediaFile, error)
}
