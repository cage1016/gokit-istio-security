package postgres

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jmoiron/sqlx"

	storageErrors "github.com/cage1016/gokit-istio-security/internal/app/authz/storage"
	"github.com/cage1016/gokit-istio-security/internal/app/authz/storage/postgres"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
)

const notifyPolicychange = "NOTIFY policychange;"

type pg struct {
	db       Database
	log      log.Logger
	connInfo string
}

// New instantiates a PostgreSQL implementation of user
// repository.
func New(_ context.Context, cfg postgres.Config, logger log.Logger) (storageV1.Storage, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		level.Error(logger).Log("method", "postgres.Connect", "err", err)
		return nil, err
	}

	return &pg{NewDatabase(db), logger, cfg.ToURL()}, nil
}

func (p pg) GetUserRoles(ctx context.Context, organizationID, userID string) ([]*storageV1.UserRole, error) {
	sql := `select opa_user_roles.organization_id, opa_user_roles.store_id, opa_roles.name, opa_roles.id 
			from opa_user_roles inner join opa_roles on opa_roles.id = opa_user_roles.role_id where user_id = $1 and organization_id = $2;`
	var userRoles []*storageV1.UserRole
	err := p.db.SelectContext(ctx, &userRoles, sql, userID, organizationID)
	if err != nil {
		return nil, p.processError(err)
	}
	return userRoles, nil
}

func (p pg) GetPolicyChangeNotifier(ctx context.Context) (storageV1.PolicyChangeNotifier, error) {
	return newPolicyChangeNotifier(ctx, p.connInfo)
}

func (p pg) GetAllUserWithRoles(ctx context.Context) ([]*storageV1.PoliciesUserRole, error) {
	roles := []*storageV1.PoliciesUserRole{}

	err := p.db.SelectContext(ctx, &roles, "select * from policies_roles")
	if err != nil {
		return nil, p.processError(err)
	}
	return roles, nil
}

func (p pg) GetAllRolesWithPermission(ctx context.Context) ([]*storageV1.PoliciesRolePermission, error) {
	rolePermissions := []*storageV1.PoliciesRolePermission{}
	err := p.db.SelectContext(ctx, &rolePermissions, `SELECT * FROM policies_permissions`)
	if err != nil {
		return nil, p.processError(err)
	}
	return rolePermissions, nil
}

func (p pg) GetRole(ctx context.Context, roleID string) (*storageV1.Role, error) {
	tx := p.db.MustBeginTx(ctx, nil)
	result, tx2 := p._getRole(ctx, tx, roleID)
	err := tx2.Commit()
	if err != nil {
		return nil, storageErrors.NewTxCommitError(err)
	}
	return result, nil
}

func (p pg) _getRole(ctx context.Context, tx Tx, roleID string) (*storageV1.Role, Tx) {
	sql := `SELECT * FROM opa_roles where id=$1`
	var role storageV1.Role
	tx.GetContext(ctx, &role, sql, roleID)

	sql = `select permission_id from opa_role_permissions where role_id =$1`
	var permissionIDs []interface{}
	tx.SelectContext(ctx, &permissionIDs, sql, roleID)
	sql = `select opa_permissions.id,
				   resource.id         "resource.id",
				   resource.name       "resource.name",
				   resource.created_at "resource.created_at",
				   resource.updated_at "resource.updated_at",
				   action.id           "action.id",
				   action.name         "action.name",
				   action.description  "action.description",
				   action.created_at   "action.created_at",
				   action.updated_at   "action.updated_at"
			from opa_permissions
					 inner join opa_resources resource on opa_permissions.resource_id = resource.id
					 inner join opa_actions action on opa_permissions.action_id = action.id
			where opa_permissions.id in (?)`
	var rolePermissions []*storageV1.RolePermission
	query, args, _ := sqlx.In(sql, permissionIDs)
	query = tx.Rebind(query)
	tx.SelectContext(ctx, &rolePermissions, query, args...)
	role.RolePermissions = rolePermissions
	return &role, tx
}

func (p pg) ListRoles(ctx context.Context) ([]*storageV1.Role, error) {
	sql := `SELECT * FROM opa_roles`
	var roles []*storageV1.Role
	if err := p.db.SelectContext(ctx, &roles, sql); err != nil {
		return nil, err
	}
	for _, role := range roles {
		sql = `select permission_id from opa_role_permissions where role_id =$1`
		var permissionIDs []string
		if err := p.db.SelectContext(ctx, &permissionIDs, sql, role.ID); err != nil {
			return nil, err
		}

		sql = `select opa_permissions.id,
					   resource.id         "resource.id",
					   resource.name       "resource.name",
					   resource.created_at "resource.created_at",
					   resource.updated_at "resource.updated_at",
					   action.id           "action.id",
					   action.name         "action.name",
					   action.description  "action.description",
					   action.created_at   "action.created_at",
					   action.updated_at   "action.updated_at"
				from opa_permissions
						 inner join opa_resources resource on opa_permissions.resource_id = resource.id
						 inner join opa_actions action on opa_permissions.action_id = action.id
				where opa_permissions.id in (?)`

		var rolePermissions []*storageV1.RolePermission
		query, args, _ := sqlx.In(sql, permissionIDs)
		query = p.db.Rebind(query)
		p.db.SelectContext(ctx, &rolePermissions, query, args...)
		role.RolePermissions = rolePermissions
	}
	return roles, nil
}

func (p *pg) processError(err error) error {
	level.Error(p.log).Log("err", fmt.Sprintf("%v", err))
	err = postgres.ProcessError(err)
	if err == storageErrors.ErrDatabase {
		level.Warn(p.log).Log("unknown_error_type_from_database", err)
	}
	return err
}
