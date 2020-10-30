package v1

type Permission struct {
	ID         string `db:"id"`
	ActionID   string `db:"action_id"`
	ResourceID string `db:"resource_id"`
}
