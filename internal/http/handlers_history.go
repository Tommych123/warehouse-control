package http

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"warehouse/internal/domain"
	"warehouse/internal/service"
)

type HistoryHandler struct {
	history *service.HistoryService
}

func NewHistoryHandler(history *service.HistoryService) *HistoryHandler {
	return &HistoryHandler{history: history}
}

func (h *HistoryHandler) ListByItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		itemID, err := parseID(chi.URLParam(r, "id"))
		if err != nil {
			Fail(w, http.StatusBadRequest, "invalid id")
			return
		}

		filter, err := parseHistoryFilter(r)
		if err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}

		includeChanges := r.URL.Query().Get("includeChanges") == "1"

		entries, err := h.history.ListByItem(r.Context(), itemID, filter, includeChanges)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to load history")
			return
		}

		JSON(w, http.StatusOK, entries)
	}
}

func (h *HistoryHandler) ExportCSV() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		itemID, err := parseID(chi.URLParam(r, "id"))
		if err != nil {
			Fail(w, http.StatusBadRequest, "invalid id")
			return
		}

		filter, err := parseHistoryFilter(r)
		if err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}

		entries, err := h.history.ListByItem(r.Context(), itemID, filter, false)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to load history")
			return
		}

		var buf bytes.Buffer
		cw := csv.NewWriter(&buf)
		_ = cw.Write([]string{"id", "item_id", "action", "actor", "actor_role", "changed_at", "old_data", "new_data"})
		for _, e := range entries {
			actor := ""
			if e.Actor != nil {
				actor = *e.Actor
			}
			role := ""
			if e.ActorRole != nil {
				role = *e.ActorRole
			}
			oldStr := compactJSON(e.OldData)
			newStr := compactJSON(e.NewData)
			_ = cw.Write([]string{
				itoa64(e.ID),
				itoa64(e.ItemID),
				e.Action,
				actor,
				role,
				e.ChangedAt.Format(time.RFC3339),
				oldStr,
				newStr,
			})
		}
		cw.Flush()

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=history_item_"+itoa64(itemID)+".csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}
}

func parseHistoryFilter(r *http.Request) (domain.HistoryFilter, error) {
	q := r.URL.Query()
	var f domain.HistoryFilter

	if from := strings.TrimSpace(q.Get("from")); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return domain.HistoryFilter{}, errBad("from must be RFC3339")
		}
		f.From = &t
	}
	if to := strings.TrimSpace(q.Get("to")); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return domain.HistoryFilter{}, errBad("to must be RFC3339")
		}
		f.To = &t
	}
	if user := strings.TrimSpace(q.Get("user")); user != "" {
		f.User = &user
	}
	if action := strings.TrimSpace(q.Get("action")); action != "" {
		f.Action = &action
	}

	return f, nil
}

type badErr string

func (e badErr) Error() string { return string(e) }
func errBad(s string) error    { return badErr(s) }
