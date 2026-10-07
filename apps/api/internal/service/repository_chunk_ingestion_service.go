package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

type RepositoryChunkIngestionService interface {
	Ingest(
		ctx context.Context,
		snapshotID string,
	) error
}

type RepositoryChunkIngestionServiceImpl struct {
	fileRepo  repository.RepositoryFileRepository
	chunkRepo repository.RepositoryChunkRepository
	chunker   *RepositoryChunker
}

func NewRepositoryChunkIngestionService(
	fileRepo repository.RepositoryFileRepository,
	chunkRepo repository.RepositoryChunkRepository,
	chunker *RepositoryChunker,
) RepositoryChunkIngestionService {
	return &RepositoryChunkIngestionServiceImpl{
		fileRepo:  fileRepo,
		chunkRepo: chunkRepo,
		chunker:   chunker,
	}
}

// Ingest creates deterministic chunks for every file in a snapshot.
// Existing chunk indexes are preserved so a retry resumes after a partial failure
// instead of creating duplicate chunks.
func (s *RepositoryChunkIngestionServiceImpl) Ingest(
	ctx context.Context,
	snapshotID string,
) error {
	files, err := s.fileRepo.ListBySnapshotID(ctx, snapshotID)
	if err != nil {
		return fmt.Errorf("list repository files for chunking: %w", err)
	}

	for _, file := range files {
		if err := s.ingestFile(ctx, file); err != nil {
			return err
		}
	}

	return nil
}

func (s *RepositoryChunkIngestionServiceImpl) ingestFile(
	ctx context.Context,
	file *domain.RepositoryFile,
) error {
	chunks := s.chunker.Chunk(file)
	for _, chunk := range chunks {
		chunk.ID = uuid.NewString()
	}

	if err := s.chunkRepo.ReplaceByFileID(ctx, file.ID, chunks); err != nil {
		return fmt.Errorf(
			"replace repository chunks for file %q: %w",
			file.Path,
			err,
		)
	}
	return nil
}

var _ RepositoryChunkIngestionService = (*RepositoryChunkIngestionServiceImpl)(nil)
