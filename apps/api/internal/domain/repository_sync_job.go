package domain

import "time"

const (
	RepositorySyncJobStatusPending   = "pending"
	RepositorySyncJobStatusRunning   = "running"
	RepositorySyncJobStatusSucceeded = "succeeded"
	RepositorySyncJobStatusFailed    = "failed"
)

type RepositorySyncJob struct {
	ID             string
	RepositoryID   string
	Status         string
	Attempts       int
	FailureCode    *string
	FailureMessage *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	CompletedAt    *time.Time
}
