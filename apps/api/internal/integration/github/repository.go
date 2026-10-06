package github

import "context"

type RepositoryTreeEntry struct {
	Path string
	Mode string
	Type string
	SHA  string
	Size int64
}

type RepositoryTree struct {
	SHA       string
	Entries   []RepositoryTreeEntry
	Truncated bool
}

type RepositoryBlob struct {
	SHA      string
	Content  string
	Encoding string
	Size     int64
}

type RepositoryClient interface {
	GetLatestCommitSHA(
		ctx context.Context,
		installationID string,
		owner string,
		repository string,
		branch string,
	) (string, error)

	GetTree(
		ctx context.Context,
		installationID string,
		owner string,
		repository string,
		treeSHA string,
	) (*RepositoryTree, error)

	GetBlob(
		ctx context.Context,
		installationID string,
		owner string,
		repository string,
		blobSHA string,
	) (*RepositoryBlob, error)
}
