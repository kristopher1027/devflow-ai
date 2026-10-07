package service

import (
	"context"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

// RepositorySyncWithIngestionService wraps a RepositorySyncService so that
// every successful sync also stores the snapshot's files.
//
// Ingestion runs even when the snapshot already existed: Ingest skips files
// that are already stored, so a retry after a half-finished ingest resumes
// where it stopped, and a fully ingested snapshot is a cheap no-op.
type RepositorySyncWithIngestionService struct {
	sync      RepositorySyncService
	ingestion RepositoryIngestionService
	chunkIngestion RepositoryChunkIngestionService
}

func NewRepositorySyncWithIngestionService(
	sync RepositorySyncService,
	ingestion RepositoryIngestionService,
	chunkIngestion ...RepositoryChunkIngestionService,
) RepositorySyncService {
	var chunker RepositoryChunkIngestionService
	if len(chunkIngestion) > 0 {
		chunker = chunkIngestion[0]
	}

	return &RepositorySyncWithIngestionService{
		sync:          sync,
		ingestion:     ingestion,
		chunkIngestion: chunker,
	}
}

func (s *RepositorySyncWithIngestionService) Sync(
	ctx context.Context,
	repositoryID string,
) (*domain.RepositorySnapshot, error) {
	snapshot, err := s.sync.Sync(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	if err := s.ingestion.Ingest(ctx, snapshot.ID); err != nil {
		return nil, err
	}

	if s.chunkIngestion != nil {
		if err := s.chunkIngestion.Ingest(ctx, snapshot.ID); err != nil {
			return nil, err
		}
	}

	return snapshot, nil
}

var _ RepositorySyncService = (*RepositorySyncWithIngestionService)(nil)
