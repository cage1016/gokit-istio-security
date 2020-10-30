package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/engine"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
	"github.com/cage1016/gokit-istio-security/internal/pkg/errors"
	"github.com/go-kit/kit/log"
)

var (
	ErrMalformedEntity = errors.New("malformed entity specification")
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(AuthzService) AuthzService

// Service describes a service that adds things together
// Implement yor service methods methods.
// e.x: Foo(ctx context.Context, s string)(rs string, err error)
type AuthzService interface {
	// [expose=false]
	IsAuthorizedReq(ctx context.Context, user string, path string, method string) (isAuthorized bool, err error)
	// [method=get,expose=true,router=roles/:id]
	GetRole(ctx context.Context, roleID string) (res *storageV1.Role, err error)
	// [method=get,expose=true,router=roles]
	ListRoles(ctx context.Context) (items []*storageV1.Role, err error)
}

// the concrete implementation of service interface
type stubAuthzService struct {
	logger          log.Logger
	store           storageV1.Storage
	engine          engine.Engine
	policyRefresher PolicyRefresher
}

// New return a new instance of the service.
// If you want to add service middleware this is the place to put them.
func New(ctx context.Context, store storageV1.Storage, e engine.Engine, policyRefresher PolicyRefresher, logger log.Logger) (s AuthzService) {
	var svc AuthzService
	{
		stubSvc := &stubAuthzService{logger: logger, store: store, engine: e, policyRefresher: policyRefresher}
		if err := stubSvc.updateEngineStore(ctx); err != nil {
			return nil
		}
		svc = stubSvc
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

func (au *stubAuthzService) updateEngineStore(ctx context.Context) (e0 error) {
	policiesRoles := make(map[string]interface{})
	{
		var userRoles []*storageV1.PoliciesUserRole
		var err error
		if userRoles, err = au.store.GetAllUserWithRoles(ctx); err != nil {
			return err
		}
		for _, u := range userRoles {
			policiesRoles[u.OrganizationIDStoreIDUserID] = strings.Split(u.RoleNames, ",")
		}
	}

	policiesPermissions := make(map[string]interface{})
	{
		type tempStruct struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		var rolePermissions []*storageV1.PoliciesRolePermission
		var err error
		if rolePermissions, err = au.store.GetAllRolesWithPermission(ctx); err != nil {
			return err
		}
		ms := make(map[string][]*tempStruct)
		for _, rolePermission := range rolePermissions {
			t := &tempStruct{Method: rolePermission.Action, Path: rolePermission.Resource}
			if _, ok := ms[rolePermission.RoleName]; !ok {
				ms[rolePermission.RoleName] = []*tempStruct{t}
			} else {
				ms[rolePermission.RoleName] = append(ms[rolePermission.RoleName], t)
			}
		}
		for k, v := range ms {
			policiesPermissions[k] = v
		}
	}

	input := map[string]interface{}{"userRoles": policiesRoles, "rolePermissions": policiesPermissions}

	enc, _ := json.Marshal(input)
	fmt.Println(string(enc))

	return au.engine.SetUserRolesAndPermissions(ctx, input)
}

// Implement the business logic of IsAuthorizedReq
func (au *stubAuthzService) IsAuthorizedReq(ctx context.Context, user string, path string, method string) (isAuthorized bool, err error) {
	return au.engine.IsAuthorized(ctx, user, path, method)
}

// Implement the business logic of GetRole
func (au *stubAuthzService) GetRole(ctx context.Context, roleID string) (res *storageV1.Role, err error) {
	return au.store.GetRole(ctx, roleID)
}

// Implement the business logic of ListRoles
func (au *stubAuthzService) ListRoles(ctx context.Context) (items []*storageV1.Role, err error) {
	return au.store.ListRoles(ctx)
}