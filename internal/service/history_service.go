package service

import (
	"context"
	"encoding/json"
	"reflect"

	"warehouse/internal/domain"
	"warehouse/internal/repo"
)

type HistoryService struct {
	repo *repo.HistoryRepo
}

func NewHistoryService(r *repo.HistoryRepo) *HistoryService {
	return &HistoryService{repo: r}
}

func (s *HistoryService) ListByItem(ctx context.Context, itemID int64, filter domain.HistoryFilter, includeChanges bool) ([]domain.HistoryEntry, error) {
	entries, err := s.repo.ListByItem(ctx, itemID, filter)
	if err != nil {
		return nil, err
	}

	if !includeChanges {
		return entries, nil
	}

	for i := range entries {
		e := &entries[i]
		if e.Action != "update" {
			continue
		}

		oldMap, _ := asMap(e.OldData)
		newMap, _ := asMap(e.NewData)
		if oldMap == nil || newMap == nil {
			continue
		}

		e.Changes = diffMaps(oldMap, newMap, map[string]bool{
			"updated_at": true,
			"created_at": true,
		})
	}

	return entries, nil
}

func asMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok
}

func diffMaps(oldMap, newMap map[string]any, ignore map[string]bool) map[string]map[string]any {
	changes := map[string]map[string]any{}

	keys := map[string]bool{}
	for k := range oldMap {
		keys[k] = true
	}
	for k := range newMap {
		keys[k] = true
	}

	for k := range keys {
		if ignore != nil && ignore[k] {
			continue
		}
		ov, okO := oldMap[k]
		nv, okN := newMap[k]

		if !okO && okN {
			changes[k] = map[string]any{"from": nil, "to": nv}
			continue
		}
		if okO && !okN {
			changes[k] = map[string]any{"from": ov, "to": nil}
			continue
		}

		if !jsonDeepEqual(ov, nv) {
			changes[k] = map[string]any{"from": ov, "to": nv}
		}
	}

	if len(changes) == 0 {
		return nil
	}
	return changes
}

func jsonDeepEqual(a, b any) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aj) == string(bj)
}
