package domain

import "time"

type SSHSession struct {
	ID        string    `json:"id"`
	ProjectID int64     `json:"project_id"`
	HostID    int64     `json:"host_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
