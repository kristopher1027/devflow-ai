package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

type fakeRepositoryFileRepositoryForChunking struct {
	files []*domain.RepositoryFile
	err   error
}

func (f *fakeRepositoryFileRepositoryForChunking) Create(context.Context, *domain.RepositoryFile) error {
	return nil
}

func (f *fakeRepositoryFileRepositoryForChunking) FindByID(context.Context, string) (*domain.RepositoryFile, error) {
	return nil, nil
}

func (f *fakeRepositoryFileRepositoryForChunking) ListBySnapshotID(context.Context, string) ([]*domain.RepositoryFile, error) {
	return f.files, f.err
}

type fakeRepositoryChunkRepositoryForChunking struct {
	chunks       map[string][]*domain.RepositoryChunk
	replaceErr   error
	replaceCalls int
}

func (f *fakeRepositoryChunkRepositoryForChunking) Create(context.Context, *domain.RepositoryChunk) error {
	return nil
}

func (f *fakeRepositoryChunkRepositoryForChunking) FindByID(context.Context, string) (*domain.RepositoryChunk, error) {
	return nil, nil
}

func (f *fakeRepositoryChunkRepositoryForChunking) ListByFileID(context.Context, string) ([]*domain.RepositoryChunk, error) {
	return f.chunks["unused"], nil
}

func (f *fakeRepositoryChunkRepositoryForChunking) ReplaceByFileID(
	_ context.Context,
	fileID string,
	chunks []*domain.RepositoryChunk,
) error {
	f.replaceCalls++
	if f.replaceErr != nil {
		return f.replaceErr
	}
	if f.chunks == nil {
		f.chunks = make(map[string][]*domain.RepositoryChunk)
	}
	f.chunks[fileID] = chunks
	return nil
}

func newChunkIngestionTestFile() *domain.RepositoryFile {
	return &domain.RepositoryFile{
		ID:         "file-1",
		SnapshotID: "snapshot-1",
		Path:       "main.go",
		Content:    "aa\nbb\ncc\ndd",
	}
}

func TestRepositoryChunkIngestionCreatesCurrentChunks(t *testing.T) {
	file := newChunkIngestionTestFile()
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{}
	chunker, err := NewRepositoryChunker(10, 0)
	require.NoError(t, err)

	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)

	err = service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 1, chunkRepo.replaceCalls)
	require.Len(t, chunkRepo.chunks[file.ID], 2)
	require.Equal(t, 0, chunkRepo.chunks[file.ID][0].ChunkIndex)
	require.Equal(t, 1, chunkRepo.chunks[file.ID][1].ChunkIndex)
	require.NotEmpty(t, chunkRepo.chunks[file.ID][0].ID)
}

func TestRepositoryChunkIngestionReplacesStaleChunks(t *testing.T) {
	file := newChunkIngestionTestFile()
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: map[string][]*domain.RepositoryChunk{
			file.ID: {
				{ID: "stale", FileID: file.ID, ChunkIndex: 9, Content: "stale"},
			},
		},
	}
	chunker, err := NewRepositoryChunker(100, 0)
	require.NoError(t, err)

	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, chunker)
	err = service.Ingest(context.Background(), "snapshot-1")

	require.NoError(t, err)
	require.Equal(t, 1, chunkRepo.replaceCalls)
	require.NotEqual(t, "stale", chunkRepo.chunks[file.ID][0].ID)
	require.Len(t, chunkRepo.chunks[file.ID], 1)
}

func TestRepositoryChunkIngestionEmptyFileClearsExistingChunks(t *testing.T) {
	file := newChunkIngestionTestFile()
	file.Content = ""
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{file},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{
		chunks: map[string][]*domain.RepositoryChunk{
			file.ID: {{ID: "stale", FileID: file.ID, ChunkIndex: 0, Content: "stale"}},
		},
	}
	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, NewDefaultRepositoryChunker())

	err := service.Ingest(context.Background(), "snapshot-1")
	require.NoError(t, err)
	require.Equal(t, 1, chunkRepo.replaceCalls)
	require.Empty(t, chunkRepo.chunks[file.ID])
}

func TestRepositoryChunkIngestionPropagatesFileListError(t *testing.T) {
	expected := errors.New("file list failed")
	fileRepo := &fakeRepositoryFileRepositoryForChunking{err: expected}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{}

	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, NewDefaultRepositoryChunker())
	err := service.Ingest(context.Background(), "snapshot-1")

	require.ErrorIs(t, err, expected)
	require.Zero(t, chunkRepo.replaceCalls)
}

func TestRepositoryChunkIngestionPropagatesReplaceError(t *testing.T) {
	expected := errors.New("replace failed")
	fileRepo := &fakeRepositoryFileRepositoryForChunking{
		files: []*domain.RepositoryFile{newChunkIngestionTestFile()},
	}
	chunkRepo := &fakeRepositoryChunkRepositoryForChunking{replaceErr: expected}

	service := NewRepositoryChunkIngestionService(fileRepo, chunkRepo, NewDefaultRepositoryChunker())
	err := service.Ingest(context.Background(), "snapshot-1")

	require.ErrorIs(t, err, expected)
	require.Equal(t, 1, chunkRepo.replaceCalls)
}
