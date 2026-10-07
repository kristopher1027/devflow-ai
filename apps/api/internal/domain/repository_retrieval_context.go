package domain

type RepositoryRetrievalContext struct {
	SnapshotID     string
	Items          []*RepositoryRetrievalContextItem
	CharacterCount int
}

type RepositoryRetrievalContextItem struct {
	ChunkID       string
	FileID        string
	ChunkIndex    int
	FilePath      string
	Language      *string
	StartLine     int
	EndLine       int
	CharacterCount int
	Content       string
	Score         float64
	MatchedTerms  []string
}
