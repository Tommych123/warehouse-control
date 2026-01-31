package http

import (
	"net/http"

	"warehouse/internal/domain"
)

func RequireRoles(allowed ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFromContext(r.Context())
			if !ok {
				Fail(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if !domain.HasAnyRole(p.Role, allowed...) {
				Fail(w, http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
