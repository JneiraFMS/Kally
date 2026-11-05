package outbound

import "context"

// MediaRepository is the storage port required by the media application.
// Names are content-addressed and validated by the domain before reaching
// the adapter, so implementations only deal with opaque names.
type MediaRepository interface {
	Save(ctx context.Context, name string, content []byte) error
	Find(ctx context.Context, name string) ([]byte, error)
}
