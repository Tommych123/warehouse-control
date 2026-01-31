package repo

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"warehouse/internal/domain"
)

type HistoryRepo struct {
	pool *pgxpool.Pool
}

func NewHistoryRepo(db *DB) *HistoryRepo {
	return &HistoryRepo{pool: db.Pool}
}

func (r *HistoryRepo) ListByItem(ctx context.Context, itemID int64, f domain.HistoryFilter) ([]domain.HistoryEntry, error) {
	q := `
select id, item_id, action, actor, actor_role, changed_at, old_data, new_data
from items_history
where item_id = $1
`
	args := []any{itemID}
	idx := 2

	if f.From != nil {
		q += ` and changed_at >= $` + strconv.Itoa(idx)
		args = append(args, *f.From)
		idx++
	}
	if f.To != nil {
		q += ` and changed_at <= $` + strconv.Itoa(idx)
		args = append(args, *f.To)
		idx++
	}
	if f.User != nil && strings.TrimSpace(*f.User) != "" {
		q += ` and actor = $` + strconv.Itoa(idx)
		args = append(args, strings.TrimSpace(*f.User))
		idx++
	}
	if f.Action != nil && strings.TrimSpace(*f.Action) != "" {
		q += ` and action = $` + strconv.Itoa(idx)
		args = append(args, strings.TrimSpace(*f.Action))
		idx++
	}

	q += ` order by changed_at desc, id desc`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.HistoryEntry, 0)
	for rows.Next() {
		var e domain.HistoryEntry
		var oldBytes, newBytes []byte
		if err := rows.Scan(&e.ID, &e.ItemID, &e.Action, &e.Actor, &e.ActorRole, &e.ChangedAt, &oldBytes, &newBytes); err != nil {
			return nil, err
		}

		if len(oldBytes) > 0 {
			var m any
			if err := json.Unmarshal(oldBytes, &m); err == nil {
				e.OldData = m
			}
		}
		if len(newBytes) > 0 {
			var m any
			if err := json.Unmarshal(newBytes, &m); err == nil {
				e.NewData = m
			}
		}

		out = append(out, e)
	}
	return out, rows.Err()
}
