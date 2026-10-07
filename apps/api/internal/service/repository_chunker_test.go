package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func TestRepositoryChunkerIsDeterministicAndLineAware(t *testing.T) {
	chunker, err := NewRepositoryChunker(20, 1)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:      "file-1",
		Content: "one\ntwo\nthree\nfour\nfive",
	}

	first := chunker.Chunk(file)
	second := chunker.Chunk(file)

	require.Equal(t, first, second)
	require.Len(t, first, 2)
	require.Equal(t, 0, first[0].ChunkIndex)
	require.Equal(t, 1, first[0].StartLine)
	require.Equal(t, 3, first[0].EndLine)
	require.Equal(t, "one\ntwo\nthree", first[0].Content)
	require.Equal(t, 1, first[1].ChunkIndex)
	require.Equal(t, 3, first[1].StartLine)
	require.Equal(t, 5, first[1].EndLine)
	require.Equal(t, "three\nfour\nfive", first[1].Content)
	require.Equal(t, 13, first[0].CharacterCount)
}

func TestRepositoryChunkerHandlesOversizedSingleLine(t *testing.T) {
	chunker, err := NewRepositoryChunker(5, 0)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:      "file-1",
		Content: "123456789\nsmall",
	}

	chunks := chunker.Chunk(file)

	require.Len(t, chunks, 2)
	require.Equal(t, "123456789", chunks[0].Content)
	require.Equal(t, 9, chunks[0].CharacterCount)
	require.Equal(t, "small", chunks[1].Content)
}

func TestRepositoryChunkerCountsUnicodeRunes(t *testing.T) {
	chunker, err := NewRepositoryChunker(10, 0)
	require.NoError(t, err)

	file := &domain.RepositoryFile{
		ID:      "file-1",
		Content: "héllo\n世界",
	}

	chunks := chunker.Chunk(file)

	require.Len(t, chunks, 1)
	require.Equal(t, 8, chunks[0].CharacterCount)
}

func TestRepositoryChunkerHandlesEmptyFile(t *testing.T) {
	chunker := NewDefaultRepositoryChunker()

	require.Empty(t, chunker.Chunk(&domain.RepositoryFile{
		ID:      "file-1",
		Content: "",
	}))
	require.Empty(t, chunker.Chunk(nil))
}

func TestRepositoryChunkerPreservesAllLinesWithoutOverlap(t *testing.T) {
	chunker, err := NewRepositoryChunker(8, 0)
	require.NoError(t, err)

	lines := []string{"aa", "bb", "cc", "dd"}
	file := &domain.RepositoryFile{
		ID:      "file-1",
		Content: strings.Join(lines, "\n"),
	}

	chunks := chunker.Chunk(file)

	require.Len(t, chunks, 2)
	require.Equal(t, "aa\nbb\ncc", chunks[0].Content)
	require.Equal(t, "dd", chunks[1].Content)
	require.Equal(t, 1, chunks[0].StartLine)
	require.Equal(t, 3, chunks[0].EndLine)
	require.Equal(t, 4, chunks[1].StartLine)
	require.Equal(t, 4, chunks[1].EndLine)
}

func TestNewRepositoryChunkerRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewRepositoryChunker(0, 0)
	require.ErrorIs(t, err, ErrRepositoryChunkMaxCharactersInvalid)

	_, err = NewRepositoryChunker(10, -1)
	require.ErrorIs(t, err, ErrRepositoryChunkOverlapInvalid)
}
