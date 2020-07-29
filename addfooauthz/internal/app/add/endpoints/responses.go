package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokit-istio-security/internal/app/add/service"
	"github.com/cage1016/gokit-istio-security/internal/pkg/responses"
)

var (
	_ httptransport.Headerer = (*SumResponse)(nil)

	_ httptransport.StatusCoder = (*SumResponse)(nil)

	_ httptransport.Headerer = (*ConcatResponse)(nil)

	_ httptransport.StatusCoder = (*ConcatResponse)(nil)
)

// SumResponse collects the response values for the Sum method.
type SumResponse struct {
	Res int64 `json:"res"`
	Err error `json:"-"`
}

func (r SumResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r SumResponse) Headers() http.Header {
	return http.Header{}
}

func (r SumResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}

// ConcatResponse collects the response values for the Concat method.
type ConcatResponse struct {
	Res string `json:"res"`
	Err error  `json:"-"`
}

func (r ConcatResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r ConcatResponse) Headers() http.Header {
	return http.Header{}
}

func (r ConcatResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}
