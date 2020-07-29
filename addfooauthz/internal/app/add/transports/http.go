package transports

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/rs/cors"
	"google.golang.org/grpc/status"

	"github.com/cage1016/gokit-istio-security/internal/app/add/endpoints"
	"github.com/cage1016/gokit-istio-security/internal/pkg/errors"
	"github.com/cage1016/gokit-istio-security/internal/pkg/jwt"
	"github.com/cage1016/gokit-istio-security/internal/pkg/responses"
)

const (
	contentType string = "application/json"
)

// NewHTTPHandler returns a handler that makes a set of endpoints available on
// predefined paths.
func NewHTTPHandler(endpoints endpoints.Endpoints, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) http.Handler { // Zipkin HTTP Server Trace can either be instantiated per endpoint with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit endpoint as ServerOption.
	// In the latter case, the operation name will be the endpoint's http method.
	// We demonstrate a global tracing service here.
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer)

	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(httpEncodeError),
		httptransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	m := bone.New()
	m.Post("/sum", httptransport.NewServer(
		endpoints.SumEndpoint,
		decodeHTTPSumRequest,
		encodeJSONResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Sum", logger), jwt.HTTPToContext()))...,
	))
	m.Post("/concat", httptransport.NewServer(
		endpoints.ConcatEndpoint,
		decodeHTTPConcatRequest,
		encodeJSONResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Concat", logger), jwt.HTTPToContext()))...,
	))
	return cors.AllowAll().Handler(m)
}

// decodeHTTPSumRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body. Primarily useful in a server.
func decodeHTTPSumRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.SumRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// decodeHTTPConcatRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body. Primarily useful in a server.
func decodeHTTPConcatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.ConcatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func httpEncodeError(_ context.Context, err error, w http.ResponseWriter) {
	code := http.StatusInternalServerError
	var message string
	var errs []errors.Errors
	w.Header().Set("Content-Type", contentType)
	if s, ok := status.FromError(err); !ok {
		// HTTP
		switch errorVal := err.(type) {
		case errors.Error:
			switch {
			// TODO write your own custom error check here
			}

			if errorVal.Msg() != "" {
				message, errs = errorVal.Msg(), errorVal.Errors()
			}
		default:
			switch err {
			case io.ErrUnexpectedEOF, io.EOF:
				code = http.StatusBadRequest
			default:
				switch err.(type) {
				case *json.SyntaxError, *json.UnmarshalTypeError:
					code = http.StatusBadRequest
				}
			}

			errs = errors.FromError(err.Error())
			message = errs[0].Message
		}
	} else {
		// GRPC
		code = HTTPStatusFromCode(s.Code())
		errs = errors.FromError(s.Message())
		message = errs[0].Message
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(responses.ErrorRes{responses.ErrorResItem{code, message, errs}})
}

func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if headerer, ok := response.(httptransport.Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}
	code := http.StatusOK
	if sc, ok := response.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
	}
	w.WriteHeader(code)
	if code == http.StatusNoContent {
		return nil
	}

	if ar, ok := response.(responses.Responser); ok {
		return json.NewEncoder(w).Encode(ar.Response())
	}

	return json.NewEncoder(w).Encode(response)
}
