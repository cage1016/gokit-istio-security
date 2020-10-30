package endpoints

import (
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/cage1016/gokit-istio-security/internal/app/authz/service"
	storageV1 "github.com/cage1016/gokit-istio-security/internal/app/authz/storage/v1"
	"github.com/cage1016/gokit-istio-security/internal/pkg/responses"
)

var (
	_ httptransport.Headerer = (*IsAuthorizedReqResponse)(nil)

	_ httptransport.StatusCoder = (*IsAuthorizedReqResponse)(nil)

	_ httptransport.Headerer = (*GetRoleResponse)(nil)

	_ httptransport.StatusCoder = (*GetRoleResponse)(nil)

	_ httptransport.Headerer = (*ListRolesResponse)(nil)

	_ httptransport.StatusCoder = (*ListRolesResponse)(nil)
)

// IsAuthorizedReqResponse collects the response values for the IsAuthorizedReq method.
type IsAuthorizedReqResponse struct {
	IsAuthorized bool  `json:"isAuthorized"`
	Err          error `json:"-"`
}

func (r IsAuthorizedReqResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r IsAuthorizedReqResponse) Headers() http.Header {
	return http.Header{}
}

func (r IsAuthorizedReqResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}

// GetRoleResponse collects the response values for the GetRole method.
type GetRoleResponse struct {
	Res *storageV1.Role `json:"res"`
	Err error           `json:"-"`
}

func (r GetRoleResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r GetRoleResponse) Headers() http.Header {
	return http.Header{}
}

func (r GetRoleResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r.Res}
}

// ListRolesResponse collects the response values for the ListRoles method.
type ListRolesResponse struct {
	Items []*storageV1.Role `json:"items"`
	Err   error             `json:"-"`
}

func (r ListRolesResponse) StatusCode() int {
	return http.StatusOK // TBA
}

func (r ListRolesResponse) Headers() http.Header {
	return http.Header{}
}

func (r ListRolesResponse) Response() interface{} {
	return responses.DataRes{APIVersion: service.Version, Data: r}
}