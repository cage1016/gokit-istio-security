package endpoints

type Request interface {
	validate() error
}

// IsAuthorizedReqRequest collects the request parameters for the IsAuthorizedReq method.
type IsAuthorizedReqRequest struct {
	User   string `json:"user"`
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (r IsAuthorizedReqRequest) validate() error {
	return nil // TBA
}

// GetRoleRequest collects the request parameters for the GetRole method.
type GetRoleRequest struct {
	RoleID string `json:"roleId"`
}

func (r GetRoleRequest) validate() error {
	return nil // TBA
}

// ListRolesRequest collects the request parameters for the ListRoles method.
type ListRolesRequest struct {
}

func (r ListRolesRequest) validate() error {
	return nil // TBA
}