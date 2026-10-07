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
	if len(chunks) == 0 {
		return nil
	}

	existing, err := s.chunkRepo.ListByFileID(ctx, file.ID)
	if err != nil {
		return fmt.Errorf(
			"list repository chunks for file %q: %w",
			file.Path,
			err,
		)
	}

	existingIndexes := make(map[int]struct{}, len(existing))
	for _, chunk := range existing {
		existingIndexes[chunk.ChunkIndex] = struct{}{}
	}

	for _, chunk := range chunks {
		if _, exists := existingIndexes[chunk.ChunkIndex]; exists {
			continue
		}

		chunk.ID = uuid.NewString()
		if err := s.chunkRepo.Create(ctx, chunk); err != nil {
			return fmt.Errorf(
				"create repository chunk for file %q at index %d: %w",
				file.Path,
				chunk.ChunkIndex,
				err,
			)
		}
	}

	return nil
}

var _ RepositoryChunkIngestionService = (*RepositoryChunkIngestionServiceImpl)(nil)
