package transports

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/endpoints"
	"github.com/cage1016/gokit-istio-security/internal/app/authz/service"
	storageErrors "github.com/cage1016/gokit-istio-security/internal/app/authz/storage"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
	"github.com/cage1016/gokit-istio-security/internal/pkg/errors"
	"github.com/cage1016/gokit-istio-security/internal/pkg/jwt"
	pb "github.com/cage1016/gokit-istio-security/pb/authz"
)

type grpcServer struct {
	isAuthorizedReq grpctransport.Handler
	getRole         grpctransport.Handler
	listRoles       grpctransport.Handler
}

func (s *grpcServer) IsAuthorizedReq(ctx context.Context, req *pb.IsAuthorizedReqRequest) (rep *pb.IsAuthorizedReqResponse, err error) {
	_, rp, err := s.isAuthorizedReq.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcEncodeError(errors.Cast(err))
	}
	rep = rp.(*pb.IsAuthorizedReqResponse)
	return rep, nil
}

func (s *grpcServer) GetRole(ctx context.Context, req *pb.GetRoleRequest) (rep *pb.GetRoleResponse, err error) {
	_, rp, err := s.getRole.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcEncodeError(errors.Cast(err))
	}
	rep = rp.(*pb.GetRoleResponse)
	return rep, nil
}

func (s *grpcServer) ListRoles(ctx context.Context, req *pb.ListRolesRequest) (rep *pb.ListRolesResponse, err error) {
	_, rp, err := s.listRoles.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcEncodeError(errors.Cast(err))
	}
	rep = rp.(*pb.ListRolesResponse)
	return rep, nil
}

// MakeGRPCServer makes a set of endpoints available as a gRPC server.
func MakeGRPCServer(endpoints endpoints.Endpoints, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) (req pb.AuthzServer) { // Zipkin GRPC Server Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit gRPC server as a
	// ServerOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path if used in combination with the Go kit gRPC Interceptor.
	//
	// In this example, we demonstrate a global Zipkin tracing service with
	// Go kit gRPC Interceptor.
	zipkinServer := zipkin.GRPCServerTrace(zipkinTracer)

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	return &grpcServer{
		isAuthorizedReq: grpctransport.NewServer(
			endpoints.IsAuthorizedReqEndpoint,
			decodeGRPCIsAuthorizedReqRequest,
			encodeGRPCIsAuthorizedReqResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "IsAuthorizedReq", logger), jwt.GRPCToContext()))...,
		),

		getRole: grpctransport.NewServer(
			endpoints.GetRoleEndpoint,
			decodeGRPCGetRoleRequest,
			encodeGRPCGetRoleResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "GetRole", logger), jwt.GRPCToContext()))...,
		),

		listRoles: grpctransport.NewServer(
			endpoints.ListRolesEndpoint,
			decodeGRPCListRolesRequest,
			encodeGRPCListRolesResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "ListRoles", logger), jwt.GRPCToContext()))...,
		),
	}
}

// decodeGRPCIsAuthorizedReqRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC request to a user-domain request. Primarily useful in a server.
func decodeGRPCIsAuthorizedReqRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.IsAuthorizedReqRequest)
	return endpoints.IsAuthorizedReqRequest{User: req.User, Path: req.Path, Method: req.Method}, nil
}

// encodeGRPCIsAuthorizedReqResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC reply. Primarily useful in a server.
func encodeGRPCIsAuthorizedReqResponse(_ context.Context, grpcReply interface{}) (res interface{}, err error) {
	reply := grpcReply.(endpoints.IsAuthorizedReqResponse)
	return &pb.IsAuthorizedReqResponse{IsAuthorized: reply.IsAuthorized}, grpcEncodeError(errors.Cast(reply.Err))
}

// decodeGRPCGetRoleRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC request to a user-domain request. Primarily useful in a server.
func decodeGRPCGetRoleRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.GetRoleRequest)
	return endpoints.GetRoleRequest{RoleID: req.RoleId}, nil
}

// encodeGRPCGetRoleResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC reply. Primarily useful in a server.
func encodeGRPCGetRoleResponse(_ context.Context, grpcReply interface{}) (res interface{}, err error) {
	reply := grpcReply.(endpoints.GetRoleResponse)
	var rolePermissions []*pb.RolePermission
	for _, v := range reply.Res.RolePermissions {
		rolePermissions = append(rolePermissions, &pb.RolePermission{
			ID: v.ID,
			Resource: &pb.Resource{
				Id:        v.Resource.ID,
				Name:      v.Resource.Name,
				CreatedAt: v.Resource.CreatedAt.Format(time.RFC3339),
				UpdatedAt: v.Resource.UpdatedAt.Format(time.RFC3339),
			},
			Action: &pb.Action{
				Id:          v.Action.ID,
				Name:        v.Action.Name,
				Description: v.Action.Description,
				CreatedAt:   v.Resource.CreatedAt.Format(time.RFC3339),
				UpdatedAt:   v.Resource.UpdatedAt.Format(time.RFC3339),
			},
		})
	}
	return &pb.GetRoleResponse{Res: &pb.Role{
		Id:              reply.Res.ID,
		Name:            reply.Res.Name,
		RolePermissions: rolePermissions,
		CreatedAt:       reply.Res.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       reply.Res.UpdatedAt.Format(time.RFC3339),
	}}, grpcEncodeError(errors.Cast(reply.Err))
}

// decodeGRPCListRolesRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC request to a user-domain request. Primarily useful in a server.
func decodeGRPCListRolesRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	_ = grpcReq.(*pb.ListRolesRequest)
	return endpoints.ListRolesRequest{}, nil
}

// encodeGRPCListRolesResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain response to a gRPC reply. Primarily useful in a server.
func encodeGRPCListRolesResponse(_ context.Context, grpcReply interface{}) (res interface{}, err error) {
	reply := grpcReply.(endpoints.ListRolesResponse)
	var roles []*pb.Role
	for _, v := range reply.Items {
		var rolePermissions []*pb.RolePermission
		for _, v := range v.RolePermissions {
			rolePermissions = append(rolePermissions, &pb.RolePermission{
				ID: v.ID,
				Resource: &pb.Resource{
					Id:        v.Resource.ID,
					Name:      v.Resource.Name,
					CreatedAt: v.Resource.CreatedAt.Format(time.RFC3339),
					UpdatedAt: v.Resource.UpdatedAt.Format(time.RFC3339),
				},
				Action: &pb.Action{
					Id:          v.Action.ID,
					Name:        v.Action.Name,
					Description: v.Action.Description,
					CreatedAt:   v.Resource.CreatedAt.Format(time.RFC3339),
					UpdatedAt:   v.Resource.UpdatedAt.Format(time.RFC3339),
				},
			})
		}
		roles = append(roles, &pb.Role{
			Id:              v.ID,
			Name:            v.Name,
			RolePermissions: rolePermissions,
			CreatedAt:       v.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       v.UpdatedAt.Format(time.RFC3339),
		})
	}
	return &pb.ListRolesResponse{Items: roles}, grpcEncodeError(errors.Cast(reply.Err))
}

// NewGRPCClient returns an AddService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) service.AuthzService { // Zipkin GRPC Client Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing client can be instantiated
	// without an operation name and fed to each Go kit client as ClientOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path.
	//
	// In this example, we demonstrace a global tracing client.
	zipkinClient := zipkin.GRPCClientTrace(zipkinTracer)

	// global client middlewares
	options := []grpctransport.ClientOption{
		zipkinClient,
	}

	// The IsAuthorizedReq endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var isAuthorizedReqEndpoint endpoint.Endpoint
	{
		isAuthorizedReqEndpoint = grpctransport.NewClient(
			conn,
			"pb.Authz",
			"IsAuthorizedReq",
			encodeGRPCIsAuthorizedReqRequest,
			decodeGRPCIsAuthorizedReqResponse,
			pb.IsAuthorizedReqResponse{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger), jwt.ContextToGRPC()))...,
		).Endpoint()
		isAuthorizedReqEndpoint = opentracing.TraceClient(otTracer, "IsAuthorizedReq")(isAuthorizedReqEndpoint)
	}

	// The GetRole endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var getRoleEndpoint endpoint.Endpoint
	{
		getRoleEndpoint = grpctransport.NewClient(
			conn,
			"pb.Authz",
			"GetRole",
			encodeGRPCGetRoleRequest,
			decodeGRPCGetRoleResponse,
			pb.GetRoleResponse{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger), jwt.ContextToGRPC()))...,
		).Endpoint()
		getRoleEndpoint = opentracing.TraceClient(otTracer, "GetRole")(getRoleEndpoint)
	}

	// The ListRoles endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var listRolesEndpoint endpoint.Endpoint
	{
		listRolesEndpoint = grpctransport.NewClient(
			conn,
			"pb.Authz",
			"ListRoles",
			encodeGRPCListRolesRequest,
			decodeGRPCListRolesResponse,
			pb.ListRolesResponse{},
			append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger), jwt.ContextToGRPC()))...,
		).Endpoint()
		listRolesEndpoint = opentracing.TraceClient(otTracer, "ListRoles")(listRolesEndpoint)
	}

	return endpoints.Endpoints{
		IsAuthorizedReqEndpoint: isAuthorizedReqEndpoint,
		GetRoleEndpoint:         getRoleEndpoint,
		ListRolesEndpoint:       listRolesEndpoint,
	}
}

// encodeGRPCIsAuthorizedReqRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain IsAuthorizedReq request to a gRPC IsAuthorizedReq request. Primarily useful in a client.
func encodeGRPCIsAuthorizedReqRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(endpoints.IsAuthorizedReqRequest)
	return &pb.IsAuthorizedReqRequest{User: req.User, Path: req.Path, Method: req.Method}, nil
}

// decodeGRPCIsAuthorizedReqResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC IsAuthorizedReq reply to a user-domain IsAuthorizedReq response. Primarily useful in a client.
func decodeGRPCIsAuthorizedReqResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.IsAuthorizedReqResponse)
	return endpoints.IsAuthorizedReqResponse{IsAuthorized: reply.IsAuthorized}, nil
}

// encodeGRPCGetRoleRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain GetRole request to a gRPC GetRole request. Primarily useful in a client.
func encodeGRPCGetRoleRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(endpoints.GetRoleRequest)
	return &pb.GetRoleRequest{RoleId: req.RoleID}, nil
}

// decodeGRPCGetRoleResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC GetRole reply to a user-domain GetRole response. Primarily useful in a client.
func decodeGRPCGetRoleResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.GetRoleResponse)
	var rolePermission []*storageV1.RolePermission
	for _, v := range reply.Res.RolePermissions {
		rolePermission = append(rolePermission, &storageV1.RolePermission{
			ID: v.GetID(),
			Resource: &storageV1.Resource{
				ID:   v.Resource.GetId(),
				Name: v.Resource.GetName(),
				UpdatedAt: func() time.Time {
					t, _ := time.Parse(time.RFC3339, v.Resource.UpdatedAt)
					return t
				}(),
				CreatedAt: func() time.Time {
					t, _ := time.Parse(time.RFC3339, v.Resource.CreatedAt)
					return t
				}(),
			},
			Action: &storageV1.Action{
				ID:          v.Action.GetId(),
				Name:        v.Action.GetName(),
				Description: v.Action.GetDescription(),
				UpdatedAt: func() time.Time {
					t, _ := time.Parse(time.RFC3339, v.Action.UpdatedAt)
					return t
				}(),
				CreatedAt: func() time.Time {
					t, _ := time.Parse(time.RFC3339, v.Action.CreatedAt)
					return t
				}(),
			},
		})
	}
	return endpoints.GetRoleResponse{Res: &storageV1.Role{
		ID:              reply.Res.GetId(),
		Name:            reply.Res.GetName(),
		RolePermissions: rolePermission,
		UpdatedAt: func() time.Time {
			t, _ := time.Parse(time.RFC3339, reply.Res.UpdatedAt)
			return t
		}(),
		CreatedAt: func() time.Time {
			t, _ := time.Parse(time.RFC3339, reply.Res.CreatedAt)
			return t
		}(),
	}}, nil
}

// encodeGRPCListRolesRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain ListRoles request to a gRPC ListRoles request. Primarily useful in a client.
func encodeGRPCListRolesRequest(_ context.Context, request interface{}) (interface{}, error) {
	_ = request.(endpoints.ListRolesRequest)
	return &pb.ListRolesRequest{}, nil
}

// decodeGRPCListRolesResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC ListRoles reply to a user-domain ListRoles response. Primarily useful in a client.
func decodeGRPCListRolesResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.ListRolesResponse)
	var roles []*storageV1.Role
	for _, v := range reply.Items {
		var rolePermission []*storageV1.RolePermission
		for _, v := range v.RolePermissions {
			rolePermission = append(rolePermission, &storageV1.RolePermission{
				ID: v.GetID(),
				Resource: &storageV1.Resource{
					ID:   v.Resource.GetId(),
					Name: v.Resource.GetName(),
					UpdatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, v.Resource.UpdatedAt)
						return t
					}(),
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, v.Resource.CreatedAt)
						return t
					}(),
				},
				Action: &storageV1.Action{
					ID:          v.Action.GetId(),
					Name:        v.Action.GetName(),
					Description: v.Action.GetDescription(),
					UpdatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, v.Action.UpdatedAt)
						return t
					}(),
					CreatedAt: func() time.Time {
						t, _ := time.Parse(time.RFC3339, v.Action.CreatedAt)
						return t
					}(),
				},
			})
		}
		roles = append(roles, &storageV1.Role{
			ID:              v.GetId(),
			Name:            v.GetName(),
			RolePermissions: rolePermission,
			UpdatedAt: func() time.Time {
				t, _ := time.Parse(time.RFC3339, v.UpdatedAt)
				return t
			}(),
			CreatedAt: func() time.Time {
				t, _ := time.Parse(time.RFC3339, v.CreatedAt)
				return t
			}(),
		})
	}
	return endpoints.ListRolesResponse{Items: roles}, nil
}

func grpcEncodeError(err errors.Error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		return status.Error(st.Code(), st.Message())
	}

	switch {
	case errors.Contains(err, storageErrors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Contains(err, storageErrors.ErrConflict):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Contains(err, storageErrors.ErrDatabase):
		return status.Error(codes.Internal, err.Error())
	case errors.Contains(err, jwt.ErrXJWTContextMissing):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
