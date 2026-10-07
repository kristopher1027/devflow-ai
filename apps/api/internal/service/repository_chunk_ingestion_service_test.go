package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type fakeRepositoryFileRepositoryForChunking struct {
	files []*domain.RepositoryFile
	err   error
}

func (f *fakeRepositoryFileRepositoryForChunking) Create(
	context.Context,
	*domain.RepositoryFile,
) error {
	return nil
}

func (f *fakeRepositoryFileRepositoryForChunking) FindByID(
	context.Context,
	string,
) (*domain.RepositoryFile, error) {
	return nil, nil
}

func (f *fakeRepositoryFileRepositoryForChunking) ListBySnapshotID(
	context.Context,
	string,
) ([]*domain.RepositoryFile, error) {
	return f.files, f.err
}

type fakeRepositoryChunkRepositoryForChunking struct {
	chunks       map[string][]*domain.RepositoryChunk
	createErr    error
	listErr      error
	createCalls  int
}

func (f *fakeRepositoryChunkRepositoryForChunking) Create(
	_ context.Context,
	chunk *domain.RepositoryChunk,
) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	if f.chunks == nil {
		f.chunks = make(map[string][]*domain.RepositoryChunk)
	}
	f.chunks[chunk.FileID] = append(f.chunks[chunk.FileID], chunk)
	return nil
}

func (f *fakeRepositoryChunkRepositoryForChunking) FindByID(
	_ context.Context,
	id string,
) (*domain.RepositoryChunk, error) {
	for _, chunks := range f.chunks {
		for _, chunk := range chunks {
			if chunk.ID == id {
				return chunk, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func (f *fakeRepositoryChunkRepositoryForChunking) ListByFileID(
	_ context.Context,
	fileID string,
) ([]*domain.RepositoryChunk, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.chunks[fileID], nil
}

func TestRepositoryChunkIngestionCreatesChunksForSnapshotFiles(t *testing.T) {
	chunker, err := NewRepositoryChunker(10, 0)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:         "file-1",
		SnapshotID: "snapshot-1",
		Path:       "main.go",
		Content:    "aa\nbb\ncc\ndd",
		CreatedAt:  time.Now().UTC(),
	}

	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: make(map[string][]*domain.RepositoryChunk),
	}
	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)

	err = service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 2, chunkRepo.createCalls)
	require.Len(t, chunkRepo.chunks[file.ID], 2)
	require.Equal(t, 0, chunkRepo.chunks[file.ID][0].ChunkIndex)
	require.Equal(t, 1, chunkRepo.chunks[file.ID][1].ChunkIndex)
	require.NotEmpty(t, chunkRepo.chunks[file.ID][0].ID)
	require.NotEmpty(t, chunkRepo.chunks[file.ID][1].ID)
}

func TestRepositoryChunkIngestionIsIdempotent(t *testing.T) {
	chunker, err := NewRepositoryChunker(10, 0)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:      "file-1",
		SnapshotID: "snapshot-1",
		Path:    "main.go",
		Content: "aa\nbb\ncc\ndd",
	}

	existing := &domain.RepositoryChunk{
		ID:         "existing",
		FileID:     file.ID,
		ChunkIndex: 0,
		StartLine:  1,
		EndLine:    3,
		CharacterCount: 8,
		Content:    "aa\nbb\ncc",
	}

	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: map[string][]*domain.RepositoryChunk{
			file.ID: {existing},
		},
	}
	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)

	err = service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 1, chunkRepo.createCalls)
	require.Len(t, chunkRepo.chunks[file.ID], 2)
	require.Equal(t, 0, chunkRepo.chunks[file.ID][0].ChunkIndex)
	require.Equal(t, 1, chunkRepo.chunks[file.ID][1].ChunkIndex)
	require.Equal(t, "existing", chunkRepo.chunks[file.ID][0].ID)
}

func TestRepositoryChunkIngestionResumesAfterPartialFailure(t *testing.T) {
	chunker, err := NewRepositoryChunker(10, 0)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:      "file-1",
		SnapshotID: "snapshot-1",
		Path:    "main.go",
		Content: "aa\nbb\ncc\ndd",
	}

	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: map[string][]*domain.RepositoryChunk{
			file.ID: {
				{
					ID:         "existing",
					FileID:     file.ID,
					ChunkIndex: 0,
					StartLine:  1,
					EndLine:    3,
					CharacterCount: 8,
					Content:    "aa\nbb\ncc",
				},
			},
		},
	}
	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)

	err = service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 1, chunkRepo.createCalls)
	require.Len(t, chunkRepo.chunks[file.ID], 2)
	require.Equal(t, 1, chunkRepo.chunks[file.ID][1].ChunkIndex)
}

func TestRepositoryChunkIngestionSkipsEmptyFiles(t *testing.T) {
	chunker := NewDefaultRepositoryChunker()
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{
			{ID: "file-1", SnapshotID: "snapshot-1", Path: "empty.go", Content: ""},
		},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: make(map[string][]*domain.RepositoryChunk),
	}
	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)

	err := service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 0, chunkRepo.createCalls)
}

func TestRepositoryChunkIngestionPropagatesFileListError(t *testing.T) {
	expected := errors.New("file list failed")
	fileRepo := &fakeRepositoryFileRepositoryForChunking{err: expected}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{}
	service := NewRepositoryChunkIngestionService(
		fileRepo,
		chunkRepo,
		NewDefaultRepositoryChunker(),
	)

	err := service.Ingest(context.Background(), "snapshot-1")
	require.ErrorIs(t, err, expected)
	require.Equal(t, 0, chunkRepo.createCalls)
}

func TestRepositoryChunkIngestionPropagatesChunkListError(t *testing.T) {
	expected := errors.New("chunk list failed")
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{
			{ID: "file-1", SnapshotID: "snapshot-1", Path: "main.go", Content: "package main"},
		},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		listErr: expected,
	}
	service := NewRepositoryChunkIngestionService(
		fileRepo,
		chunkRepo,
		NewDefaultRepositoryChunker(),
	)

	err := service.Ingest(context.Background(), "snapshot-1")
	require.ErrorIs(t, err, expected)
	require.Equal(t, 0, chunkRepo.createCalls)
}

func TestRepositoryChunkIngestionPropagatesCreateError(t *testing.T) {
	expected := errors.New("create failed")
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{
			{ID: "file-1", SnapshotID: "snapshot-1", Path: "main.go", Content: "package main"},
		},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		createErr: expected,
	}
	service := NewRepositoryChunkIngestionService(
		fileRepo,
		chunkRepo,
		NewDefaultRepositoryChunker(),
	)

	err := service.Ingest(context.Background(), "snapshot-1")
	require.ErrorIs(t, err, expected)
	require.Equal(t, 1, chunkRepo.createCalls)
}
