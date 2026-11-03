package inbound

import (
	"context"

	"grz/internal/application/dto"
	domainjournal "grz/internal/domain/journal"
)

// JournalApplication exposes the daily-journal use cases to inbound
// adapters. The date of the day ("today") is supplied by the inbound
// adapter so use cases stay deterministic.
type JournalApplication interface {
	Page(ctx context.Context, today string) (dto.JournalPage, error)
	Entry(ctx context.Context, date string) (dto.JournalEntryPage, error)
	Create(ctx context.Context, input dto.JournalInput) (domainjournal.Entry, error)
	Update(ctx context.Context, input dto.JournalInput) (domainjournal.Entry, error)
	Delete(ctx context.Context, date string) error
}
