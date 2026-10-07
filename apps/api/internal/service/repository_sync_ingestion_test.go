package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type stubSyncService struct {
	snapshot *domain.RepositorySnapshot
	err      error
	calls    int
}

func (s *stubSyncService) Sync(
	ctx context.Context,
	repositoryID string,
) (*domain.RepositorySnapshot, error) {
	s.calls++
	return s.snapshot, s.err
}

type stubChunkIngestionService struct {
	err        error
	snapshotID string
	calls      int
}

func (s *stubChunkIngestionService) Ingest(
	ctx context.Context,
	snapshotID string,
) error {
	s.calls++
	s.snapshotID = snapshotID
	return s.err
}

type stubIngestionService struct {
	err        error
	snapshotID string
	calls      int
}

func (s *stubIngestionService) Ingest(
	ctx context.Context,
	snapshotID string,
) error {
	s.calls++
	s.snapshotID = snapshotID
	return s.err
}

func TestSyncWithIngestionIngestsTheSyncedSnapshot(t *testing.T) {
	snapshot := &domain.RepositorySnapshot{ID: "snap-1"}
	syncer := &stubSyncService{snapshot: snapshot}
	ingester := &stubIngestionService{}

	got, err := NewRepositorySyncWithIngestionService(syncer, ingester, &stubChunkIngestionService{}).
		Sync(context.Background(), "repo-1")

	require.NoError(t, err)
	require.Same(t, snapshot, got)
	require.Equal(t, 1, ingester.calls)
	require.Equal(t, "snap-1", ingester.snapshotID)
}

func TestSyncWithIngestionSkipsIngestionWhenSyncFails(t *testing.T) {
	syncer := &stubSyncService{err: errors.New("sync failed")}
	ingester := &stubIngestionService{}

	got, err := NewRepositorySyncWithIngestionService(syncer, ingester).
		Sync(context.Background(), "repo-1")

	require.EqualError(t, err, "sync failed")
	require.Nil(t, got)
	require.Zero(t, ingester.calls)
}

func TestSyncWithIngestionReturnsIngestionError(t *testing.T) {
	syncer := &stubSyncService{snapshot: &domain.RepositorySnapshot{ID: "snap-1"}}
	ingester := &stubIngestionService{err: errors.New("ingest failed")}

	got, err := NewRepositorySyncWithIngestionService(syncer, ingester).
		Sync(context.Background(), "repo-1")

	require.EqualError(t, err, "ingest failed")
	require.Nil(t, got)
}

func TestSyncWithIngestionRunsChunkIngestionAfterFileIngestion(t *testing.T) {
	snapshot := &domain.RepositorySnapshot{ID: "snap-1"}
	syncer := &stubSyncService{snapshot: snapshot}
	ingester := &stubIngestionService{}
	chunker := &stubChunkIngestionService{}

	got, err := NewRepositorySyncWithIngestionService(syncer, ingester, chunker).
		Sync(context.Background(), "repo-1")

	require.NoError(t, err)
	require.Same(t, snapshot, got)
	require.Equal(t, 1, ingester.calls)
	require.Equal(t, 1, chunker.calls)
	require.Equal(t, "snap-1", chunker.snapshotID)
}

func TestSyncWithIngestionReturnsChunkIngestionError(t *testing.T) {
	syncer := &stubSyncService{snapshot: &domain.RepositorySnapshot{ID: "snap-1"}}
	ingester := &stubIngestionService{}
	chunker := &stubChunkIngestionService{err: errors.New("chunk ingest failed")}

	got, err := NewRepositorySyncWithIngestionService(syncer, ingester, chunker).
		Sync(context.Background(), "repo-1")

	require.EqualError(t, err, "chunk ingest failed")
	require.Nil(t, got)
}
