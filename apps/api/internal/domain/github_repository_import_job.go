package domain

import "time"

const (
	GitHubRepositoryImportJobStatusPending   = "pending"
	GitHubRepositoryImportJobStatusRunning   = "running"
	GitHubRepositoryImportJobStatusSucceeded = "succeeded"
	GitHubRepositoryImportJobStatusFailed    = "failed"
	GitHubRepositoryImportJobFailureCodeWorkerShutdown = "worker_shutdown"
)

type GitHubRepositoryImportJob struct {
	ID             string
	RequesterID    string
	ProjectID      string
	Status         string
	Attempts       int
	FailureCode    *string
	FailureMessage *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
}
