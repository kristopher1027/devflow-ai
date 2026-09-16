package domain

import "time"

type Project struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
