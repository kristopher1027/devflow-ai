package domain

import "time"

type Repository struct {
	ID            string
	ProjectID     string
	Provider      string
	ExternalID    string
	Owner         string
	Name          string
	FullName      string
	DefaultBranch string
	HTMLURL       string
	CloneURL      string
	IsPrivate     bool
	SyncStatus    string
	LastSyncedAt  *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
