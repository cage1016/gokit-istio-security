package engine

import "context"

// Engine abstracts different decision engines.
//
// Currently, a full-featured engine needs to support both v1 and v2. These are
// split to allow for a better decoupling in wiring up things -- i.e., the v2
// authz and policy implementations only require V2Authorizer and V2Writer,
// respectively. Our OPA implementation of engine.Engine supports all of those
// interfaces, and can thus be plugged into either of those constructors.
type Engine interface {
	Authorizer
	Writer
}

type Authorizer interface {
	IsAuthorized(ctx context.Context, user, path, method string) (bool, error)
}

// Writer is the interface for writing policies to a decision engine
type Writer interface {
	SetUserRolesAndPermissions(context.Context, map[string]interface{}) error
}
