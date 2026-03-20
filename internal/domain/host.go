package domain

import "time"

type Host struct {
	ID        string       `json:"id"`
	ProjectID string       `json:"project_id"`
	Name      string       `json:"name"`
	Address   string       `json:"address"`
	Port      int          `json:"port"`
	Username  string       `json:"username"`
	AuthKind  HostAuthKind `json:"auth_kind"`
	Password  string       `json:"-"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type HostAuthKind string

const (
	HostAuthNone     HostAuthKind = "none"
	HostAuthPassword HostAuthKind = "password"
	HostAuthAgent    HostAuthKind = "agent"
)
