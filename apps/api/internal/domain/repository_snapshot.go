package domain

import "time"

type RepositorySnapshot struct {
	ID           string
	RepositoryID string
	CommitSHA    string
	Branch       string
	CreatedAt    time.Time
}
