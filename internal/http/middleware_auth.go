package http

import (
	"context"
	"net/http"
	"strings"

	"warehouse/internal/auth"
	"warehouse/internal/domain"
)

type ctxKey string

const principalKey ctxKey = "principal"

type Principal struct {
	Username string
	Role     domain.Role
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	v := ctx.Value(principalKey)
	p, ok := v.(Principal)
	return p, ok
}

func RequireAuth(jwtMgr *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" {
				Fail(w, http.StatusUnauthorized, "missing Authorization header")
				return
			}

			parts := strings.SplitN(h, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				Fail(w, http.StatusUnauthorized, "invalid Authorization header format")
				return
			}

			claims, err := jwtMgr.Parse(strings.TrimSpace(parts[1]))
			if err != nil {
				Fail(w, http.StatusUnauthorized, "invalid token")
				return
			}

			role, ok := domain.ParseRole(claims.Role)
			if !ok {
				Fail(w, http.StatusUnauthorized, "invalid role in token")
				return
			}

			p := Principal{Username: claims.Username, Role: role}
			ctx := context.WithValue(r.Context(), principalKey, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
