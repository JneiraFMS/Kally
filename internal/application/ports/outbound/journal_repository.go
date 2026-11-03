package outbound

import (
	"context"

	domainjournal "grz/internal/domain/journal"
)

// JournalRepository is the persistence port required by the journal
// application. Entries live in the same storage directory as KBs but are
// identified by date; the adapter owns the file naming convention.
type JournalRepository interface {
	Save(ctx context.Context, entry domainjournal.Entry) error
	Exists(ctx context.Context, date string) (bool, error)
	ByDate(ctx context.Context, date string) (*domainjournal.Entry, error)
	Entries(ctx context.Context) ([]domainjournal.Entry, error)
	Delete(ctx context.Context, date string) error
}
