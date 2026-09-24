package domain

import "time"

type RepositoryFile struct {
	ID         string
	SnapshotID string
	Path       string
	SizeBytes  int64
	Language   *string
	Content    string
	CreatedAt  time.Time
}
