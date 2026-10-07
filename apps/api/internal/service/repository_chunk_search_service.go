package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrRepositoryChunkSearchQueryRequired = errors.New("repository chunk search query is required")
	ErrRepositoryChunkSearchLimitInvalid  = errors.New("repository chunk search limit must be positive")
)

const DefaultRepositoryChunkSearchLimit = 10

type RepositoryChunkSearchService interface {
	Search(
		ctx context.Context,
		snapshotID string,
		query string,
		limit int,
	) ([]*domain.RepositoryChunkSearchResult, error)
}

type RepositoryChunkSearchServiceImpl struct {
	chunkRepo repository.RepositoryChunkRepository
}

func NewRepositoryChunkSearchService(
	chunkRepo repository.RepositoryChunkRepository,
) RepositoryChunkSearchService {
	return &RepositoryChunkSearchServiceImpl{chunkRepo: chunkRepo}
}

func (s *RepositoryChunkSearchServiceImpl) Search(
	ctx context.Context,
	snapshotID string,
	query string,
	limit int,
) ([]*domain.RepositoryChunkSearchResult, error) {
	if strings.TrimSpace(snapshotID) == "" {
		return nil, fmt.Errorf("snapshot id is required")
	}
	if strings.TrimSpace(query) == "" {
		return nil, ErrRepositoryChunkSearchQueryRequired
	}
	if limit <= 0 {
		return nil, ErrRepositoryChunkSearchLimitInvalid
	}

	results, err := s.chunkRepo.SearchBySnapshotID(
		ctx,
		snapshotID,
		strings.TrimSpace(query),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search repository chunks: %w", err)
	}
	return results, nil
}

var _ RepositoryChunkSearchService = (*RepositoryChunkSearchServiceImpl)(nil)
