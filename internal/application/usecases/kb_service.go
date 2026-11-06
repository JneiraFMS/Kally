package usecases

import (
	"context"
	"fmt"

	"grz/internal/application"
	"grz/internal/application/dto"
	"grz/internal/application/ports/inbound"
	"grz/internal/application/ports/outbound"
	domainkb "grz/internal/domain/kb"
	"grz/internal/kbfile"
)

// KBService orchestrates knowledge-base use cases through the repository port.
type KBService struct {
	repo outbound.KBRepository
}

// NewKBService injects the repository required by the KB use cases.
func NewKBService(repo outbound.KBRepository) *KBService {
	return &KBService{repo: repo}
}

// List returns all knowledge bases, filtered by the provided tags (AND:
// every tag must be present). Tag filter values are normalized (case,
// whitespace, duplicates) and never rejected.
func (s *KBService) List(ctx context.Context, tags ...string) ([]domainkb.KB, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	tags = domainkb.NormalizeTagFilter(tags)
	switch len(tags) {
	case 1:
		items = domainkb.FilterByTag(items, tags[0])
	case 0:
		// sans filtre : toutes les KB
	default:
		items = domainkb.FilterByAllTags(items, tags...)
	}
	return items, nil
}

// Get retrieves one knowledge base by ID.
func (s *KBService) Get(ctx context.Context, id string) (*domainkb.KB, error) {
	return s.repo.FindByID(ctx, id)
}

// Detail retrieves a knowledge base and its tag-related knowledge bases.
func (s *KBService) Detail(ctx context.Context, id string) (dto.KBDetail, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return dto.KBDetail{}, err
	}
	item := domainkb.ByID(items, id)
	if item == nil {
		return dto.KBDetail{}, fmt.Errorf("%w: %s", domainkb.ErrNotFound, id)
	}
	return dto.KBDetail{
		KB:        *item,
		Related:   domainkb.Related(items, *item),
		Backlinks: domainkb.Backlinks(items, id),
		Neighbors: domainkb.Neighbors(items, id),
		KnownIDs:  domainkb.IDs(items),
	}, nil
}

// Create validates and persists a new knowledge base, refusing duplicate IDs.
func (s *KBService) Create(ctx context.Context, input dto.KBInput) (domainkb.KB, error) {
	item, err := toDomain(input)
	if err != nil {
		return domainkb.KB{}, err
	}
	exists, err := s.repo.Exists(ctx, item.ID)
	if err != nil {
		return domainkb.KB{}, err
	}
	if exists {
		return domainkb.KB{}, fmt.Errorf("%w: %s", application.ErrAlreadyExists, item.ID)
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domainkb.KB{}, err
	}
	return item, nil
}

// Update validates and persists an existing knowledge base.
func (s *KBService) Update(ctx context.Context, input dto.KBInput) (domainkb.KB, error) {
	item, err := toDomain(input)
	if err != nil {
		return domainkb.KB{}, err
	}
	if err := s.repo.Save(ctx, item); err != nil {
		return domainkb.KB{}, err
	}
	return item, nil
}

// Delete removes a knowledge base by ID.
func (s *KBService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// SerializeLocal validates the input (same domain rules as a shared KB) and
// returns the canonical .md content, ready to be written by the client
// script into the user's local library folder. No server persistence, and
// no uniqueness check: local and shared worlds are watertight.
func (s *KBService) SerializeLocal(input dto.KBInput) (string, error) {
	item, err := toDomain(input)
	if err != nil {
		return "", err
	}
	return kbfile.Serialize(item), nil
}

// ParseLocal analyzes the .md content of a file read from the local
// library. The fields come back disaggregated for the form and the client
// detail card; the ID is provided by the file name.
func (s *KBService) ParseLocal(raw string) (dto.KBParsed, error) {
	item := kbfile.Parse("", raw)
	return dto.KBParsed{
		Erreur:     item.Erreur,
		Notes:      item.Notes,
		Resolution: item.Resolution,
		Tags:       item.Tags,
	}, nil
}

// LocalKnownIDs exposes the IDs of shared KBs, used to render [[...]] wiki
// links of a local file (a link targeting a shared ID points to the detail
// page; otherwise it becomes a dead link). Local files never enter this
// set: the two worlds remain watertight.
func (s *KBService) LocalKnownIDs(ctx context.Context) ([]string, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return domainkb.IDs(items), nil
}

func toDomain(input dto.KBInput) (domainkb.KB, error) {
	item, err := domainkb.NewKB(input.ID, input.Erreur, input.Notes, input.Resolution, input.Tags)
	if err != nil {
		return domainkb.KB{}, err
	}
	return item, nil
}

var _ inbound.KBApplication = (*KBService)(nil)
