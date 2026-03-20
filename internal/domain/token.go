package domain

import "time"

type APIToken struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TokenProject struct {
	TokenID   string `json:"token_id"`
	ProjectID string `json:"project_id"`
}
