package transports

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/rs/cors"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/endpoints"
	storageErrors "github.com/cage1016/gokit-istio-security/internal/app/authz/storage"
	"github.com/cage1016/gokit-istio-security/internal/pkg/errors"
	"github.com/cage1016/gokit-istio-security/internal/pkg/jwt"
	"github.com/cage1016/gokit-istio-security/internal/pkg/responses"
)

const (
	contentType string = "application/json"
)

// ShowAuthz godoc
// @Summary GetRole
// @Description TODO
// @Tags TODO
// @Accept json
// @Produce json
// @Router /roles/:id [get]
func GetRoleHandler(m *bone.Mux, endpoints endpoints.Endpoints, options []httptransport.ServerOption, otTracer stdopentracing.Tracer, logger log.Logger) {
	m.Get("/roles/:id", httptransport.NewServer(
		endpoints.GetRoleEndpoint,
		decodeHTTPGetRoleRequest,
		responses.EncodeJSONResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetRole", logger), jwt.HTTPToContext()))...,
	))
}

// ShowAuthz godoc
// @Summary ListRoles
// @Description TODO
// @Tags TODO
// @Accept json
// @Produce json
// @Router /roles [get]
func ListRolesHandler(m *bone.Mux, endpoints endpoints.Endpoints, options []httptransport.ServerOption, otTracer stdopentracing.Tracer, logger log.Logger) {
	m.Get("/roles", httptransport.NewServer(
		endpoints.ListRolesEndpoint,
		decodeHTTPListRolesRequest,
		responses.EncodeJSONResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "ListRoles", logger), jwt.HTTPToContext()))...,
	))
}

// NewHTTPHandler returns a handler that makes a set of endpoints available on
// predefined paths.
func NewHTTPHandler(endpoints endpoints.Endpoints, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) http.Handler { // Zipkin HTTP Server Trace can either be instantiated per endpoint with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit endpoint as ServerOption.
	// In the latter case, the operation name will be the endpoint's http method.
	// We demonstrate a global tracing service here.
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(responses.ErrorEncodeJSONResponse(CustomErrorEncoder)),
		httptransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	m := bone.New()
	GetRoleHandler(m, endpoints, options, otTracer, logger)
	ListRolesHandler(m, endpoints, options, otTracer, logger)
	return cors.AllowAll().Handler(m)
}

// decodeHTTPGetRoleRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body. Primarily useful in a server.
func decodeHTTPGetRoleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.GetRoleRequest
	req.RoleID = bone.GetValue(r, "id")
	return req, nil
}

// decodeHTTPListRolesRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body. Primarily useful in a server.
func decodeHTTPListRolesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.ListRolesRequest
	return req, nil
}

func CustomErrorEncoder(errorVal errors.Error) (code int) {
	switch {
	case errors.Contains(errorVal, storageErrors.ErrNotFound):
		code = http.StatusNotFound
	case errors.Contains(errorVal, storageErrors.ErrConflict):
		code = http.StatusBadRequest
	case errors.Contains(errorVal, storageErrors.ErrDatabase):
		code = http.StatusInternalServerError
	case errors.Contains(errorVal, jwt.ErrXJWTContextMissing):
		code = http.StatusForbidden
	}
	return
}
