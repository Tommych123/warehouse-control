package domain

import "strings"

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleViewer  Role = "viewer"
)

func ParseRole(s string) (Role, bool) {
	switch Role(strings.ToLower(strings.TrimSpace(s))) {
	case RoleAdmin:
		return RoleAdmin, true
	case RoleManager:
		return RoleManager, true
	case RoleViewer:
		return RoleViewer, true
	default:
		return "", false
	}
}

func (r Role) String() string { return string(r) }

func HasAnyRole(current Role, allowed ...Role) bool {
	for _, a := range allowed {
		if current == a {
			return true
		}
	}
	return false
}
