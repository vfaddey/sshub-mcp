package domain

import "time"

type SSHSession struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	HostID    string    `json:"host_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
