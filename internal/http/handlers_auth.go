package http

import (
	"net/http"
	"strings"
	"time"

	"warehouse/internal/auth"
	"warehouse/internal/domain"
)

type loginRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func LoginHandler(jwtMgr *auth.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := DecodeJSON(r, &req); err != nil {
			Fail(w, http.StatusBadRequest, "invalid json")
			return
		}

		req.Username = strings.TrimSpace(req.Username)
		if req.Username == "" {
			Fail(w, http.StatusBadRequest, "username is required")
			return
		}

		role, ok := domain.ParseRole(req.Role)
		if !ok {
			Fail(w, http.StatusBadRequest, "invalid role")
			return
		}

		token, err := jwtMgr.Generate(req.Username, role.String(), 24*time.Hour)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to generate token")
			return
		}

		JSON(w, http.StatusOK, loginResponse{Token: token})
	}
}

func MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		JSON(w, http.StatusOK, map[string]any{
			"username": p.Username,
			"role":     p.Role,
		})
	}
}
