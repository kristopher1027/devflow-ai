package domain

import "time"

type WorkspaceMember struct {
	WorkspaceID string
	UserID      string
	Role        string
	CreatedAt   time.Time
}