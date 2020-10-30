package endpoints

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/service"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
)

// Endpoints collects all of the endpoints that compose the authz service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
type Endpoints struct {
	IsAuthorizedReqEndpoint endpoint.Endpoint
	GetRoleEndpoint         endpoint.Endpoint
	ListRolesEndpoint       endpoint.Endpoint
}

// New return a new instance of the endpoint that wraps the provided service.
func New(svc service.AuthzService, logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) (ep Endpoints) {
	var isAuthorizedReqEndpoint endpoint.Endpoint
	{
		method := "isAuthorizedReq"
		isAuthorizedReqEndpoint = MakeIsAuthorizedReqEndpoint(svc)
		isAuthorizedReqEndpoint = opentracing.TraceServer(otTracer, method)(isAuthorizedReqEndpoint)
		isAuthorizedReqEndpoint = zipkin.TraceEndpoint(zipkinTracer, method)(isAuthorizedReqEndpoint)
		isAuthorizedReqEndpoint = LoggingMiddleware(log.With(logger, "method", method))(isAuthorizedReqEndpoint)
		ep.IsAuthorizedReqEndpoint = isAuthorizedReqEndpoint
	}

	var getRoleEndpoint endpoint.Endpoint
	{
		method := "getRole"
		getRoleEndpoint = MakeGetRoleEndpoint(svc)
		getRoleEndpoint = opentracing.TraceServer(otTracer, method)(getRoleEndpoint)
		getRoleEndpoint = zipkin.TraceEndpoint(zipkinTracer, method)(getRoleEndpoint)
		getRoleEndpoint = LoggingMiddleware(log.With(logger, "method", method))(getRoleEndpoint)
		ep.GetRoleEndpoint = getRoleEndpoint
	}

	var listRolesEndpoint endpoint.Endpoint
	{
		method := "listRoles"
		listRolesEndpoint = MakeListRolesEndpoint(svc)
		listRolesEndpoint = opentracing.TraceServer(otTracer, method)(listRolesEndpoint)
		listRolesEndpoint = zipkin.TraceEndpoint(zipkinTracer, method)(listRolesEndpoint)
		listRolesEndpoint = LoggingMiddleware(log.With(logger, "method", method))(listRolesEndpoint)
		ep.ListRolesEndpoint = listRolesEndpoint
	}

	return ep
}

// MakeIsAuthorizedReqEndpoint returns an endpoint that invokes IsAuthorizedReq on the service.
// Primarily useful in a server.
func MakeIsAuthorizedReqEndpoint(svc service.AuthzService) (ep endpoint.Endpoint) {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(IsAuthorizedReqRequest)
		if err := req.validate(); err != nil {
			return IsAuthorizedReqResponse{}, err
		}
		isAuthorized, err := svc.IsAuthorizedReq(ctx, req.User, req.Path, req.Method)
		return IsAuthorizedReqResponse{IsAuthorized: isAuthorized}, err
	}
}

// IsAuthorizedReq implements the service interface, so Endpoints may be used as a service.
// This is primarily useful in the context of a client library.
func (e Endpoints) IsAuthorizedReq(ctx context.Context, user string, path string, method string) (isAuthorized bool, err error) {
	resp, err := e.IsAuthorizedReqEndpoint(ctx, IsAuthorizedReqRequest{User: user, Path: path, Method: method})
	if err != nil {
		return
	}
	response := resp.(IsAuthorizedReqResponse)
	return response.IsAuthorized, nil
}

// MakeGetRoleEndpoint returns an endpoint that invokes GetRole on the service.
// Primarily useful in a server.
func MakeGetRoleEndpoint(svc service.AuthzService) (ep endpoint.Endpoint) {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetRoleRequest)
		if err := req.validate(); err != nil {
			return GetRoleResponse{}, err
		}
		res, err := svc.GetRole(ctx, req.RoleID)
		return GetRoleResponse{Res: res}, err
	}
}

// GetRole implements the service interface, so Endpoints may be used as a service.
// This is primarily useful in the context of a client library.
func (e Endpoints) GetRole(ctx context.Context, roleID string) (res *storageV1.Role, err error) {
	resp, err := e.GetRoleEndpoint(ctx, GetRoleRequest{RoleID: roleID})
	if err != nil {
		return
	}
	response := resp.(GetRoleResponse)
	return response.Res, nil
}

// MakeListRolesEndpoint returns an endpoint that invokes ListRoles on the service.
// Primarily useful in a server.
func MakeListRolesEndpoint(svc service.AuthzService) (ep endpoint.Endpoint) {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		items, err := svc.ListRoles(ctx)
		return ListRolesResponse{Items: items}, err
	}
}

// ListRoles implements the service interface, so Endpoints may be used as a service.
// This is primarily useful in the context of a client library.
func (e Endpoints) ListRoles(ctx context.Context) (items []*storageV1.Role, err error) {
	resp, err := e.ListRolesEndpoint(ctx, ListRolesRequest{})
	if err != nil {
		return
	}
	response := resp.(ListRolesResponse)
	return response.Items, nil
}