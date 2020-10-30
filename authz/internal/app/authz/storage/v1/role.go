package v1

import (
	"time"
)

type Role struct {
	ID              string            `db:"id" json:"id"`
	Name            string            `db:"name" json:"name"`
	RolePermissions []*RolePermission `json:"role_permissions"`
	CreatedAt       time.Time         `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time         `db:"updated_at" json:"updatedAt"`
}

type RolePermission struct {
	ID       string    `db:"id" json:"id"`
	Resource *Resource `db:"resource" json:"resource"`
	Action   *Action   `db:"action" json:"action"`
}
