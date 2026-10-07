package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type fakeRepositoryChunkSearchRepository struct {
	results []*domain.RepositoryChunkSearchResult
	err     error
	query   string
	limit   int
}

func (f *fakeRepositoryChunkSearchRepository) Create(context.Context, *domain.RepositoryChunk) error { return nil }
func (f *fakeRepositoryChunkSearchRepository) FindByID(context.Context, string) (*domain.RepositoryChunk, error) { return nil, nil }
func (f *fakeRepositoryChunkSearchRepository) ListByFileID(context.Context, string) ([]*domain.RepositoryChunk, error) { return nil, nil }
func (f *fakeRepositoryChunkSearchRepository) ReplaceByFileID(context.Context, string, []*domain.RepositoryChunk) error { return nil }
func (f *fakeRepositoryChunkSearchRepository) SearchBySnapshotID(
	_ context.Context, snapshotID, query string, limit int,
) ([]*domain.RepositoryChunkSearchResult, error) {
	f.query = snapshotID + ":" + query
	f.limit = limit
	return f.results, f.err
}

func TestRepositoryChunkSearchTrimsAndDelegates(t *testing.T) {
	repo := &fakeRepositoryChunkSearchRepository{
		results: []*domain.RepositoryChunkSearchResult{{FilePath: "main.go"}},
	}
	service := NewRepositoryChunkSearchService(repo)

	results, err := service.Search(context.Background(), "snapshot-1", "  authentication  ", 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "snapshot-1:authentication", repo.query)
	require.Equal(t, 5, repo.limit)
}

func TestRepositoryChunkSearchRejectsEmptyQuery(t *testing.T) {
	service := NewRepositoryChunkSearchService(&fakeRepositoryChunkSearchRepository{})

	_, err := service.Search(context.Background(), "snapshot-1", "   ", 5)
	require.ErrorIs(t, err, ErrRepositoryChunkSearchQueryRequired)
}

func TestRepositoryChunkSearchRejectsInvalidLimit(t *testing.T) {
	service := NewRepositoryChunkSearchService(&fakeRepositoryChunkSearchRepository{})

	_, err := service.Search(context.Background(), "snapshot-1", "auth", 0)
	require.ErrorIs(t, err, ErrRepositoryChunkSearchLimitInvalid)
}

func TestRepositoryChunkSearchPropagatesRepositoryError(t *testing.T) {
	expected := errors.New("search failed")
	service := NewRepositoryChunkSearchService(
		&fakeRepositoryChunkSearchRepository{err: expected},
	)

	_, err := service.Search(context.Background(), "snapshot-1", "auth", 5)
	require.ErrorIs(t, err, expected)
}
