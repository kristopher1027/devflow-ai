package service

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/kristopher1027/devflow-ai/internal/domain"
)

const (
	DefaultRepositoryChunkMaxCharacters = 4000
	DefaultRepositoryChunkOverlapLines  = 10
)

var (
	ErrRepositoryChunkMaxCharactersInvalid = errors.New(
		"repository chunk max characters must be positive",
	)
	ErrRepositoryChunkOverlapInvalid = errors.New(
		"repository chunk overlap must be non-negative and smaller than max lines",
	)
)

type RepositoryChunker struct {
	MaxCharacters int
	OverlapLines  int
}

func NewRepositoryChunker(
	maxCharacters int,
	overlapLines int,
) (*RepositoryChunker, error) {
	if maxCharacters <= 0 {
		return nil, ErrRepositoryChunkMaxCharactersInvalid
	}
	if overlapLines < 0 {
		return nil, ErrRepositoryChunkOverlapInvalid
	}

	return &RepositoryChunker{
		MaxCharacters: maxCharacters,
		OverlapLines:  overlapLines,
	}, nil
}

func NewDefaultRepositoryChunker() *RepositoryChunker {
	chunker, err := NewRepositoryChunker(
		DefaultRepositoryChunkMaxCharacters,
		DefaultRepositoryChunkOverlapLines,
	)
	if err != nil {
		panic(err)
	}
	return chunker
}

func (c *RepositoryChunker) Chunk(
	file *domain.RepositoryFile,
) []*domain.RepositoryChunk {
	if file == nil || file.Content == "" {
		return nil
	}

	lines := strings.Split(file.Content, "
")
	chunks := make([]*domain.RepositoryChunk, 0)

	start := 0
	for start < len(lines) {
		end := c.endForLines(lines, start)
		content := strings.Join(lines[start:end], "
")

		if content == "" {
			start = end
			continue
		}

		chunks = append(chunks, &domain.RepositoryChunk{
			FileID:         file.ID,
			ChunkIndex:     len(chunks),
			StartLine:      start + 1,
			EndLine:        end,
			CharacterCount: utf8.RuneCountInString(content),
			Content:        content,
		})

		if end == len(lines) {
			break
		}

		nextStart := end - c.OverlapLines
		if nextStart <= start {
			nextStart = start + 1
		}
		start = nextStart
	}

	return chunks
}

func (c *RepositoryChunker) endForLines(
	lines []string,
	start int,
) int {
	end := start
	characters := 0

	for end < len(lines) {
		lineCharacters := utf8.RuneCountInString(lines[end])
		added := lineCharacters
		if end > start {
			added++
		}

		if end > start && characters+added > c.MaxCharacters {
			break
		}

		if end == start && lineCharacters > c.MaxCharacters {
			end++
			break
		}

		characters += added
		end++
	}

	return end
}
