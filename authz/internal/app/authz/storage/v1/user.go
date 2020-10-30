package v1

import (
	"encoding/json"
	"time"
)

type User struct {
	ID             string      `db:"id" json:"id"`
	Name           string      `db:"name" json:"name"`
	Email          string      `db:"email" json:"email"`
	UserID         string      `db:"user_id" json:"userId"`
	IsActive       bool        `db:"is_active" json:"isActive"`
	Provider       string      `db:"provider" json:"provider"`
	Avatar         string      `db:"avatar" json:"avatar"`
	CreatedAt      time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updatedAt"`
	UserRoles      []*UserRole `json:"userRoles"`
}

func (p User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		Alias
		UpdatedAt string `json:"updatedAt"`
		CreatedAt string `json:"createdAt"`
	}{
		Alias:     (Alias)(p),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
	})
}

type UserRole struct {
	ID             string `db:"id" json:"id"`
	Name           string `db:"name" json:"name"`
	StoreID        string `db:"store_id" json:"storeId"`
}
