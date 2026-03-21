package domain

import "time"

type Host struct {
	ID        int64        `json:"id"`
	ProjectID int64        `json:"project_id"`
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
