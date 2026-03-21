package domain

import "time"

type APIToken struct {
	ID        int64     `json:"id"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenProject struct {
	TokenID   int64 `json:"token_id"`
	ProjectID int64 `json:"project_id"`
}
