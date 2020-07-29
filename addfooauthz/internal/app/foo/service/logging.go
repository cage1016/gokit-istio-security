package service

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggingMiddleware struct {
	logger log.Logger `json:""`
	next   FooService `json:""`
}

// LoggingMiddleware takes a logger as a dependency
// and returns a ServiceMiddleware.
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next FooService) FooService {
		return loggingMiddleware{level.Info(logger), next}
	}
}

func (lm loggingMiddleware) Foo(ctx context.Context, s string) (res string, err error) {
	defer func() {
		lm.logger.Log("method", "Foo", "s", s, "err", err)
	}()

	return lm.next.Foo(ctx, s)
}
