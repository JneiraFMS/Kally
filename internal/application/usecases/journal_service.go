package usecases

import (
	"context"
	"fmt"

	"grz/internal/application"
	"grz/internal/application/dto"
	"grz/internal/application/ports/inbound"
	"grz/internal/application/ports/outbound"
	domainjournal "grz/internal/domain/journal"
)

// JournalService orchestrates daily-journal use cases through the
// journal-repository port. Deterministic: the date of today is injected.
type JournalService struct {
	repo outbound.JournalRepository
}

// NewJournalService injects the repository required by the journal use cases.
func NewJournalService(repo outbound.JournalRepository) *JournalService {
	return &JournalService{repo: repo}
}

// Page assembles the daily-journal page: entries newest first and whether
// today's note exists. today (ex. "2026-11-03") comes from the inbound
// adapter.
func (s *JournalService) Page(ctx context.Context, today string) (dto.JournalPage, error) {
	entries, err := s.repo.Entries(ctx)
	if err != nil {
		return dto.JournalPage{}, err
	}
	exists, err := s.repo.Exists(ctx, today)
	if err != nil {
		return dto.JournalPage{}, err
	}
	return dto.JournalPage{
		Today:       today,
		TodayExists: exists,
		Entries:     domainjournal.EntriesDesc(entries),
	}, nil
}

// Entry assembles one entry page with day navigation (previous = older,
// next = newer), refusing absent entries through the domain error.
func (s *JournalService) Entry(ctx context.Context, date string) (dto.JournalEntryPage, error) {
	entries, err := s.repo.Entries(ctx)
	if err != nil {
		return dto.JournalEntryPage{}, err
	}
	for _, entry := range entries {
		if entry.Date != date {
			continue
		}
		prev, next := domainjournal.NeighborDates(entries, date)
		return dto.JournalEntryPage{Entry: entry, Prev: prev, Next: next}, nil
	}
	return dto.JournalEntryPage{}, fmt.Errorf("%w: %s", domainjournal.ErrNotFound, date)
}

// Create validates and persists a new journal entry, refusing to overwrite
// the entry of a day that already has one.
func (s *JournalService) Create(ctx context.Context, input dto.JournalInput) (domainjournal.Entry, error) {
	entry, err := domainjournal.New(input.Date, input.Content)
	if err != nil {
		return domainjournal.Entry{}, err
	}
	exists, err := s.repo.Exists(ctx, entry.Date)
	if err != nil {
		return domainjournal.Entry{}, err
	}
	if exists {
		return domainjournal.Entry{}, fmt.Errorf("%w: %s", application.ErrAlreadyExists, entry.Date)
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return domainjournal.Entry{}, err
	}
	return entry, nil
}

// Update validates and persists the entry of a day (upsert semantics: the
// content is replaced, the original date is kept by the caller).
func (s *JournalService) Update(ctx context.Context, input dto.JournalInput) (domainjournal.Entry, error) {
	entry, err := domainjournal.New(input.Date, input.Content)
	if err != nil {
		return domainjournal.Entry{}, err
	}
	if err := s.repo.Save(ctx, entry); err != nil {
		return domainjournal.Entry{}, err
	}
	return entry, nil
}

// Delete removes a journal entry by date, propagating ErrNotFound.
func (s *JournalService) Delete(ctx context.Context, date string) error {
	return s.repo.Delete(ctx, date)
}

var _ inbound.JournalApplication = (*JournalService)(nil)
