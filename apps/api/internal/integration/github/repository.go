package github

import "context"

type RepositoryClient interface {
	GetLatestCommitSHA(
		ctx context.Context,
		owner string,
		repository string,
		branch string,
	) (string, error)
}
