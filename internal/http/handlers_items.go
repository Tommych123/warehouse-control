package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"warehouse/internal/domain"
	"warehouse/internal/repo"
	"warehouse/internal/service"
)

type ItemsHandler struct {
	items *service.ItemsService
}

func NewItemsHandler(items *service.ItemsService) *ItemsHandler {
	return &ItemsHandler{items: items}
}

type itemUpsertRequest struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Qty      int     `json:"qty"`
	Location *string `json:"location"`
}

func (h *ItemsHandler) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		items, err := h.items.List(r.Context(), search)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to list items")
			return
		}
		JSON(w, http.StatusOK, items)
	}
}

func (h *ItemsHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())

		var req itemUpsertRequest
		if err := DecodeJSON(r, &req); err != nil {
			Fail(w, http.StatusBadRequest, "invalid json")
			return
		}

		req.SKU = strings.TrimSpace(req.SKU)
		req.Name = strings.TrimSpace(req.Name)
		if req.SKU == "" || req.Name == "" {
			Fail(w, http.StatusBadRequest, "sku and name are required")
			return
		}
		if req.Qty < 0 {
			Fail(w, http.StatusBadRequest, "qty must be >= 0")
			return
		}
		if req.Location != nil {
			loc := strings.TrimSpace(*req.Location)
			req.Location = &loc
			if loc == "" {
				req.Location = nil
			}
		}

		it, err := h.items.Create(r.Context(), p.Username, p.Role.String(), domain.ItemCreate{
			SKU:      req.SKU,
			Name:     req.Name,
			Qty:      req.Qty,
			Location: req.Location,
		})
		if err != nil {
			if isUniqueViolation(err) {
				Fail(w, http.StatusConflict, "sku must be unique")
				return
			}
			Fail(w, http.StatusInternalServerError, "failed to create item")
			return
		}

		JSON(w, http.StatusCreated, it)
	}
}

func (h *ItemsHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())

		id, err := parseID(chi.URLParam(r, "id"))
		if err != nil {
			Fail(w, http.StatusBadRequest, "invalid id")
			return
		}

		var req itemUpsertRequest
		if err := DecodeJSON(r, &req); err != nil {
			Fail(w, http.StatusBadRequest, "invalid json")
			return
		}

		req.SKU = strings.TrimSpace(req.SKU)
		req.Name = strings.TrimSpace(req.Name)
		if req.SKU == "" || req.Name == "" {
			Fail(w, http.StatusBadRequest, "sku and name are required")
			return
		}
		if req.Qty < 0 {
			Fail(w, http.StatusBadRequest, "qty must be >= 0")
			return
		}
		if req.Location != nil {
			loc := strings.TrimSpace(*req.Location)
			req.Location = &loc
			if loc == "" {
				req.Location = nil
			}
		}

		it, err := h.items.Update(r.Context(), p.Username, p.Role.String(), id, domain.ItemUpdate{
			SKU:      req.SKU,
			Name:     req.Name,
			Qty:      req.Qty,
			Location: req.Location,
		})
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				Fail(w, http.StatusNotFound, "item not found")
				return
			}
			if isUniqueViolation(err) {
				Fail(w, http.StatusConflict, "sku must be unique")
				return
			}
			Fail(w, http.StatusInternalServerError, "failed to update item")
			return
		}

		JSON(w, http.StatusOK, it)
	}
}

func (h *ItemsHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())

		id, err := parseID(chi.URLParam(r, "id"))
		if err != nil {
			Fail(w, http.StatusBadRequest, "invalid id")
			return
		}

		if err := h.items.Delete(r.Context(), p.Username, p.Role.String(), id); err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				Fail(w, http.StatusNotFound, "item not found")
				return
			}
			Fail(w, http.StatusInternalServerError, "failed to delete item")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func parseID(s string) (int64, error) {
	s = strings.TrimSpace(s)
	return strconv.ParseInt(s, 10, 64)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
