package domain

import "time"

type Workspace struct {
	ID        string
	OwnerID   string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
