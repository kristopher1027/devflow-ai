package domain

import "time"

type RepositoryChunk struct {
	ID            string
	FileID        string
	ChunkIndex    int
	StartLine     int
	EndLine       int
	CharacterCount int
	Content       string
	CreatedAt     time.Time
}

type RepositoryChunkSearchResult struct {
	Chunk      *RepositoryChunk
	FilePath   string
	Language   *string
	Score      float64
	MatchedTerms []string
}
