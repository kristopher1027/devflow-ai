package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

func TestRepositoryRetrievalContextBuilderBuild(t *testing.T) {
	language := "go"

	builder, err := NewRepositoryRetrievalContextBuilder(100)
	require.NoError(t, err)

	results := []*domain.RepositoryChunkSearchResult{
		{
			Chunk: &domain.RepositoryChunk{
				ID:          "chunk-2",
				FileID:      "file-2",
				ChunkIndex: 1,
				StartLine:   20,
				EndLine:     30,
				Content:     "package service",
			},
			FilePath:     "internal/service/search.go",
			Language:     &language,
			Score:        8.5,
			MatchedTerms: []string{"service"},
		},
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "chunk-1",
				FileID:    "file-1",
				ChunkIndex: 0,
				StartLine:  1,
				EndLine:   10,
				Content:   "package domain",
			},
			FilePath: "internal/domain/chunk.go",
			Score:    5,
		},
	}

	context, err := builder.Build("snapshot-1", results)
	require.NoError(t, err)
	require.Equal(t, "snapshot-1", context.SnapshotID)
	require.Equal(t, 2, len(context.Items))
	require.Equal(t, "internal/domain/chunk.go", context.Items[0].FilePath)
	require.Equal(t, "internal/service/search.go", context.Items[1].FilePath)
	require.Equal(t, 29, context.CharacterCount)
	require.Equal(t, 8.5, context.Items[1].Score)
	require.Equal(t, []string{"service"}, context.Items[1].MatchedTerms)
	require.Equal(t, "go", *context.Items[1].Language)
}

func TestRepositoryRetrievalContextBuilderRejectsInvalidConfiguration(t *testing.T) {
	_, err := NewRepositoryRetrievalContextBuilder(0)
	require.ErrorIs(t, err, ErrRepositoryRetrievalContextMaxCharactersInvalid)
}

func TestRepositoryRetrievalContextBuilderRejectsMissingSnapshotID(t *testing.T) {
	builder, err := NewRepositoryRetrievalContextBuilder(100)
	require.NoError(t, err)

	_, err = builder.Build("  ", nil)
	require.ErrorIs(t, err, ErrRepositoryRetrievalContextSnapshotIDRequired)
}

func TestRepositoryRetrievalContextBuilderRespectsCharacterBudget(t *testing.T) {
	builder, err := NewRepositoryRetrievalContextBuilder(10)
	require.NoError(t, err)

	results := []*domain.RepositoryChunkSearchResult{
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "large",
				FileID:    "file-1",
				StartLine: 1,
				EndLine:   3,
				Content:   "12345678901",
			},
			FilePath: "a.go",
		},
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "small",
				FileID:    "file-2",
				StartLine: 1,
				EndLine:   2,
				Content:   "12345",
			},
			FilePath: "b.go",
		},
	}

	context, err := builder.Build("snapshot-1", results)
	require.NoError(t, err)
	require.Len(t, context.Items, 1)
	require.Equal(t, "b.go", context.Items[0].FilePath)
	require.Equal(t, 5, context.CharacterCount)
}

func TestRepositoryRetrievalContextBuilderDeduplicatesOverlappingChunks(t *testing.T) {
	builder, err := NewRepositoryRetrievalContextBuilder(100)
	require.NoError(t, err)

	results := []*domain.RepositoryChunkSearchResult{
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "first",
				FileID:    "file-1",
				StartLine: 1,
				EndLine:   20,
				Content:   "first",
			},
			FilePath: "internal/service/search.go",
			Score:    10,
		},
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "overlap",
				FileID:    "file-1",
				StartLine: 11,
				EndLine:   30,
				Content:   "overlap",
			},
			FilePath: "internal/service/search.go",
			Score:    9,
		},
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "next",
				FileID:    "file-1",
				StartLine: 31,
				EndLine:   40,
				Content:   "next",
			},
			FilePath: "internal/service/search.go",
			Score:    8,
		},
	}

	context, err := builder.Build("snapshot-1", results)
	require.NoError(t, err)
	require.Len(t, context.Items, 2)
	require.Equal(t, "first", context.Items[0].ChunkID)
	require.Equal(t, "next", context.Items[1].ChunkID)
}

func TestRepositoryRetrievalContextBuilderSkipsNilAndInvalidChunks(t *testing.T) {
	builder, err := NewRepositoryRetrievalContextBuilder(100)
	require.NoError(t, err)

	results := []*domain.RepositoryChunkSearchResult{
		nil,
		{Chunk: nil},
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "invalid",
				FileID:    "file-1",
				StartLine: 0,
				EndLine:   2,
				Content:   "invalid",
			},
			FilePath: "invalid.go",
		},
	}

	context, err := builder.Build("snapshot-1", results)
	require.NoError(t, err)
	require.Empty(t, context.Items)
	require.Zero(t, context.CharacterCount)
}

func TestRepositoryRetrievalContextBuilderUsesRuneCount(t *testing.T) {
	builder, err := NewRepositoryRetrievalContextBuilder(2)
	require.NoError(t, err)

	results := []*domain.RepositoryChunkSearchResult{
		{
			Chunk: &domain.RepositoryChunk{
				ID:        "unicode",
				FileID:    "file-1",
				StartLine: 1,
				EndLine:   1,
				Content:   "é🙂",
			},
			FilePath: "unicode.go",
		},
	}

	context, err := builder.Build("snapshot-1", results)
	require.NoError(t, err)
	require.Len(t, context.Items, 1)
	require.Equal(t, 2, context.CharacterCount)
	require.Equal(t, 2, context.Items[0].CharacterCount)
}
