package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/kristopher1027/devflow-ai/internal/domain"
	"github.com/kristopher1027/devflow-ai/internal/repository"
)

var (
	ErrRepositoryChunkSearchQueryRequired = errors.New("repository chunk search query is required")
	ErrRepositoryChunkSearchLimitInvalid  = errors.New("repository chunk search limit must be positive")
)

const DefaultRepositoryChunkSearchLimit = 10

type RepositoryChunkSearchService interface {
	Search(
		ctx context.Context,
		snapshotID string,
		query string,
		limit int,
	) ([]*domain.RepositoryChunkSearchResult, error)
}

type RepositoryChunkSearchServiceImpl struct {
	chunkRepo repository.RepositoryChunkRepository
}

func NewRepositoryChunkSearchService(
	chunkRepo repository.RepositoryChunkRepository,
) RepositoryChunkSearchService {
	return &RepositoryChunkSearchServiceImpl{chunkRepo: chunkRepo}
}

func (s *RepositoryChunkSearchServiceImpl) Search(
	ctx context.Context,
	snapshotID string,
	query string,
	limit int,
) ([]*domain.RepositoryChunkSearchResult, error) {
	if strings.TrimSpace(snapshotID) == "" {
		return nil, fmt.Errorf("snapshot id is required")
	}
	if strings.TrimSpace(query) == "" {
		return nil, ErrRepositoryChunkSearchQueryRequired
	}
	if limit <= 0 {
		return nil, ErrRepositoryChunkSearchLimitInvalid
	}

	fetchLimit := limit
	if fetchLimit < 3 {
		fetchLimit = 3
	}

	terms := tokenizeRepositorySearchQuery(query)
	if len(terms) == 0 {
		return nil, ErrRepositoryChunkSearchQueryRequired
	}

	results, err := s.chunkRepo.SearchBySnapshotID(
		ctx,
		snapshotID,
		strings.Join(terms, " "),
		fetchLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("search repository chunks: %w", err)
	}

	results = rankRepositoryChunkResults(results, query)
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func rankRepositoryChunkResults(
	results []*domain.RepositoryChunkSearchResult,
	query string,
) []*domain.RepositoryChunkSearchResult {
	terms := tokenizeRepositorySearchQuery(query)

	for _, result := range results {
		if result == nil || result.Chunk == nil {
			continue
		}
		result.Score = 0
		result.MatchedTerms = nil
		content := strings.ToLower(result.Chunk.Content)
		path := strings.ToLower(result.FilePath)
		for _, term := range terms {
			matched := false
			if count := strings.Count(path, term); count > 0 {
				result.Score += float64(5 * count)
				matched = true
			}
			if count := strings.Count(content, term); count > 0 {
				result.Score += float64(2 * count)
				matched = true
			}
			for _, segment := range strings.FieldsFunc(path, func(r rune) bool {
				return r == '/' || r == '\\' || r == '.' || r == '-' || r == '_' || r == ' '
			}) {
				if len(segment) < 2 {
					continue
				}
				if strings.Contains(term, segment) || strings.Contains(segment, term) {
					result.Score += 4
					matched = true
				}
			}
			if matched {
				result.MatchedTerms = append(result.MatchedTerms, term)
			}
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		left, right := results[i], results[j]
		if left == nil || left.Chunk == nil {
			return false
		}
		if right == nil || right.Chunk == nil {
			return true
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if len(left.FilePath) != len(right.FilePath) {
			return len(left.FilePath) > len(right.FilePath)
		}
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		return left.Chunk.ChunkIndex < right.Chunk.ChunkIndex
	})
	return results
}

func tokenizeRepositorySearchQuery(query string) []string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return unicode.IsSpace(r) ||
			strings.ContainsRune(".,;:!?()[]{}<>/\\\\-_+=*#\"'`", r)
	})

	seen := make(map[string]struct{})
	terms := make([]string, 0, len(fields))

	for _, field := range fields {
		runes := []rune(field)
		start := 0

		for i := 1; i < len(runes); i++ {
			prev, current := runes[i-1], runes[i]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])

			lowerToUpper := (unicode.IsLower(prev) || unicode.IsDigit(prev)) &&
				unicode.IsUpper(current)
			acronymBoundary := unicode.IsUpper(prev) &&
				unicode.IsUpper(current) && nextIsLower

			if lowerToUpper || acronymBoundary {
				addRepositorySearchTerm(string(runes[start:i]), seen, &terms)
				start = i
			}
		}

		addRepositorySearchTerm(string(runes[start:]), seen, &terms)
	}

	return terms
}

func addRepositorySearchTerm(value string, seen map[string]struct{}, terms *[]string) {
	value = strings.ToLower(value)
	if len([]rune(value)) < 2 {
		return
	}
	if _, exists := seen[value]; exists {
		return
	}
	seen[value] = struct{}{}
	*terms = append(*terms, value)
}

var _ RepositoryChunkSearchService = (*RepositoryChunkSearchServiceImpl)(nil)
