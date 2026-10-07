package service

import (
	"errors"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

var (
	ErrRepositoryRetrievalContextSnapshotIDRequired = errors.New("repository retrieval context snapshot id is required")
	ErrRepositoryRetrievalContextMaxCharactersInvalid = errors.New("repository retrieval context max characters must be positive")
)

const DefaultRepositoryRetrievalContextMaxCharacters = 12000

type RepositoryRetrievalContextBuilder interface {
	Build(
		snapshotID string,
		results []*domain.RepositoryChunkSearchResult,
	) (*domain.RepositoryRetrievalContext, error)
}

type RepositoryRetrievalContextBuilderImpl struct {
	maxCharacters int
}

func NewRepositoryRetrievalContextBuilder(
	maxCharacters int,
) (RepositoryRetrievalContextBuilder, error) {
	if maxCharacters <= 0 {
		return nil, ErrRepositoryRetrievalContextMaxCharactersInvalid
	}

	return &RepositoryRetrievalContextBuilderImpl{
		maxCharacters: maxCharacters,
	}, nil
}

func NewDefaultRepositoryRetrievalContextBuilder() RepositoryRetrievalContextBuilder {
	builder, err := NewRepositoryRetrievalContextBuilder(
		DefaultRepositoryRetrievalContextMaxCharacters,
	)
	if err != nil {
		panic(err)
	}
	return builder
}

func (b *RepositoryRetrievalContextBuilderImpl) Build(
	snapshotID string,
	results []*domain.RepositoryChunkSearchResult,
) (*domain.RepositoryRetrievalContext, error) {
	snapshotID = strings.TrimSpace(snapshotID)
	if snapshotID == "" {
		return nil, ErrRepositoryRetrievalContextSnapshotIDRequired
	}

	context := &domain.RepositoryRetrievalContext{
		SnapshotID: snapshotID,
		Items:      make([]*domain.RepositoryRetrievalContextItem, 0),
	}

	selectedRanges := make([]selectedRepositoryChunkRange, 0, len(results))
	seenChunkIDs := make(map[string]struct{})
	remaining := b.maxCharacters

	for _, result := range results {
		if result == nil || result.Chunk == nil {
			continue
		}

		chunk := result.Chunk
		if chunk.StartLine <= 0 || chunk.EndLine < chunk.StartLine {
			continue
		}

		if chunk.ID != "" {
			if _, exists := seenChunkIDs[chunk.ID]; exists {
				continue
			}
			seenChunkIDs[chunk.ID] = struct{}{}
		}

		fileKey := chunk.FileID
		if fileKey == "" {
			fileKey = result.FilePath
		}

		overlaps := false
		for _, selected := range selectedRanges {
			if selected.fileKey != fileKey {
				continue
			}
			if chunk.StartLine <= selected.endLine && chunk.EndLine >= selected.startLine {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}

		characterCount := utf8.RuneCountInString(chunk.Content)
		if characterCount > remaining {
			continue
		}

		item := &domain.RepositoryRetrievalContextItem{
			ChunkID:        chunk.ID,
			FileID:         chunk.FileID,
			ChunkIndex:     chunk.ChunkIndex,
			FilePath:      result.FilePath,
			Language:      cloneLanguage(result.Language),
			StartLine:     chunk.StartLine,
			EndLine:       chunk.EndLine,
			CharacterCount: characterCount,
			Content:       chunk.Content,
			Score:         result.Score,
			MatchedTerms:  append([]string(nil), result.MatchedTerms...),
		}

		context.Items = append(context.Items, item)
		context.CharacterCount += characterCount
		remaining -= characterCount
		selectedRanges = append(selectedRanges, selectedRepositoryChunkRange{
			fileKey:    fileKey,
			startLine:  chunk.StartLine,
			endLine:    chunk.EndLine,
		})
	}

	sort.SliceStable(context.Items, func(i, j int) bool {
		left, right := context.Items[i], context.Items[j]
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		if left.EndLine != right.EndLine {
			return left.EndLine < right.EndLine
		}
		if left.ChunkIndex != right.ChunkIndex {
			return left.ChunkIndex < right.ChunkIndex
		}
		return left.ChunkID < right.ChunkID
	})

	return context, nil
}

type selectedRepositoryChunkRange struct {
	fileKey   string
	startLine int
	endLine   int
}

func cloneLanguage(language *string) *string {
	if language == nil {
		return nil
	}
	value := *language
	return &value
}

var _ RepositoryRetrievalContextBuilder = (*RepositoryRetrievalContextBuilderImpl)(nil)

