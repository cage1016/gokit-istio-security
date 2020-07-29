package service

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
)

type loggingMiddleware struct {
	logger log.Logger   `json:""`
	next   AuthzService `json:""`
}

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next AuthzService) AuthzService {
		return loggingMiddleware{level.Info(logger), next}
	}
}

func (lm loggingMiddleware) IsAuthorizedReq(ctx context.Context, user string, path string, method string) (isAuthorized bool, err error) {
	defer func() {
		lm.logger.Log("method", "IsAuthorizedReq", "user", user, "path", path, "method", method, "err", err)
	}()

	return lm.next.IsAuthorizedReq(ctx, user, path, method)
}

func (lm loggingMiddleware) GetRole(ctx context.Context, roleID string) (res *storageV1.Role, err error) {
	defer func() {
		lm.logger.Log("method", "GetRole", "roleID", roleID, "err", err)
	}()

	return lm.next.GetRole(ctx, roleID)
}

func (lm loggingMiddleware) ListRoles(ctx context.Context) (items []*storageV1.Role, err error) {
	defer func() {
		lm.logger.Log("method", "ListRoles", "err", err)
	}()

	return lm.next.ListRoles(ctx)
}