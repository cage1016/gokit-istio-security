package rbac.authz

import data.userRoles
import data.rolePermissions

default allow = false

allow {
    some r
    roles_for_user[r]

    p := rolePermissions[r][_]
    p.method == input.method
    re_match(p.path, input.path)
}

roles_for_user[parsed] {
    some key
    parsed := userRoles[key][_]
    key == input.user
}