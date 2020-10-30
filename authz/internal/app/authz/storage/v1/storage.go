package v1

import (
	"context"
)

type Storage interface {
	policiesStorage
	ActionStorage
	ResourceStorage
	RoleStorage
	UserStorage
}

type PolicyChangeNotification struct{}

type PolicyChangeNotifier interface {
	C() <-chan PolicyChangeNotification
	Close() error
}

// opa
type policiesStorage interface {
	GetAllUserWithRoles(context.Context) ([]*PoliciesUserRole, error)
	GetAllRolesWithPermission(context.Context) ([]*PoliciesRolePermission, error)
	GetPolicyChangeNotifier(context.Context) (PolicyChangeNotifier, error)
}

type ActionStorage interface {
	//CreateAction(context.Context, *Path) (*Path, error)
	//DeleteAction(context.Context, string) error
	//UpdateAction(context.Context, *Path) (*Path, error)
	//ListActions(context.Context) ([]*Path, error)
	//GetAction(context.Context, string) (*Path, error)
}

type ResourceStorage interface {
	//CreateResource(context.Context, *Method) (*Method, error)
	//DeleteResource(context.Context, string) error
	//UpdateResource(context.Context, *Method) (*Method, error)
	//ListResources(context.Context) ([]*Method, error)
	//GetResource(context.Context, string) (*Method, error)
}

type RoleStorage interface {
	//CreateRole(ctx context.Context, role *Role) (*Role, error)
	GetRole(ctx context.Context, roleID string) (*Role, error)
	//UpdateRole(ctx context.Context, role *Role) (*Role, error)
	//DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context) ([]*Role, error)
}

type UserStorage interface {
	//SyncUser(ctx context.Context, user *User) (*User, error)
	//GetUser(ctx context.Context, userID string) (*User, error)
	//UpdateUser(ctx context.Context, userID string, user *User) (*User, error)
	//ListUsers(ctx context.Context string) ([]*User, error)
	//GetUserRoles(ctx context.Context, userID string) ([]*UserRole, error)
}
