package domain

type RepositoryAIContextPackage struct {
	SnapshotID     string
	Prompt         string
	Sources        []*RepositoryAIContextSource
	CharacterCount int
}

type RepositoryAIContextSource struct {
	ChunkID      string
	FilePath     string
	StartLine    int
	EndLine      int
	Language     *string
	Score        float64
	MatchedTerms []string
}
