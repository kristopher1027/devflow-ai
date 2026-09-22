package github

import "context"

type RepositoryClient interface {
	GetLatestCommitSHA(
		ctx context.Context,
		installationID string,
		owner string,
		repository string,
		branch string,
	) (string, error)
}