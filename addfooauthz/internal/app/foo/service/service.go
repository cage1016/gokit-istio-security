package service

import (
	"context"

	"github.com/go-kit/kit/log"

	addservice "github.com/cage1016/gokit-istio-security/internal/app/add/service"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(FooService) FooService

// Service describes a service that adds things together
// Implement yor service methods methods.
// e.x: Foo(ctx context.Context, s string)(rs string, err error)
type FooService interface {
	// [method=post,expose=true]
	Foo(ctx context.Context, s string) (res string, err error)
}

// the concrete implementation of service interface
type stubFooService struct {
	logger log.Logger `json:"logger"`
	add    addservice.AddService
}

// New return a new instance of the service.
// If you want to add service middleware this is the place to put them.
func New(add addservice.AddService, logger log.Logger) (s FooService) {
	var svc FooService
	{
		svc = &stubFooService{logger: logger, add: add}
		svc = LoggingMiddleware(logger)(svc)
	}
	return svc
}

// Implement the business logic of Foo
func (fo *stubFooService) Foo(ctx context.Context, s string) (res string, err error) {
	return fo.add.Concat(ctx, s, " bar")
}
