package usecases

import (
	"context"

	"grz/internal/application/dto"
	"grz/internal/application/ports/inbound"
	"grz/internal/application/ports/outbound"
	domainmedia "grz/internal/domain/media"
)

// mediaURLPrefix is the public path served by the media HTTP adapter.
const mediaURLPrefix = "/media/"

// MediaService orchestrates pasted-image storage through the repository port.
type MediaService struct {
	repo outbound.MediaRepository
}

// NewMediaService injects the storage port required by the media use cases.
func NewMediaService(repo outbound.MediaRepository) *MediaService {
	return &MediaService{repo: repo}
}

// Upload validates the image (format and size), stores it under its
// content-addressed name and returns the public reference.
func (s *MediaService) Upload(ctx context.Context, content []byte, mime string) (dto.Media, error) {
	name, err := domainmedia.NameFor(mime, content)
	if err != nil {
		return dto.Media{}, err
	}
	if err := s.repo.Save(ctx, name, content); err != nil {
		return dto.Media{}, err
	}
	return dto.Media{Name: name, URL: mediaURLPrefix + name}, nil
}

// Get reads a stored image by name. An invalid name is rejected before the
// repository is queried.
func (s *MediaService) Get(ctx context.Context, name string) (dto.MediaFile, error) {
	mime, err := domainmedia.ContentType(name)
	if err != nil {
		return dto.MediaFile{}, err
	}
	content, err := s.repo.Find(ctx, name)
	if err != nil {
		return dto.MediaFile{}, err
	}
	return dto.MediaFile{Name: name, Content: content, MIME: mime}, nil
}

var _ inbound.MediaApplication = (*MediaService)(nil)
