package jwt

const (
	ClaimsSub    = "sub"
	ClaimsIss    = "iss"
	ClaimsUserID = "userId"
	ClaimsRoles  = "roles"
)

type Claims struct {
	Sub    string   `json:"sub"`
	Issuer string   `json:"iss"`
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"`
}
