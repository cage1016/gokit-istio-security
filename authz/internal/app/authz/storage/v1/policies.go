package v1

type PoliciesRolePermission struct {
	RoleName string `db:"role_name"`
	Action   string `db:"action_name"`
	Resource string `db:"resource_name"`
}

type PoliciesUserRole struct {
	OrganizationIDStoreIDUserID string `db:"user_id"`
	RoleNames                   string `db:"role_names"`
}