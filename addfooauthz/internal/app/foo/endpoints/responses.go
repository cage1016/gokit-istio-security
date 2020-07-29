package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokit-istio-security/internal/app/foo/service"
	"github.com/cage1016/gokit-istio-security/internal/pkg/responses"
)

var (
	_ httptransport.Headerer = (*FooResponse)(nil)

	_ httptransport.StatusCoder = (*FooResponse)(nil)
)

// FooResponse collects the response values for the Foo method.
type FooResponse struct {
	Res string `json:"res"`
	Err error  `json:"-"`
}

func (r FooResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r FooResponse) Headers() http.Header {
	return http.Header{}
}

func (r FooResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}
