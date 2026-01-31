package domain

import "time"

type HistoryEntry struct {
	ID        int64     `json:"id"`
	ItemID    int64     `json:"item_id"`
	Action    string    `json:"action"`
	Actor     *string   `json:"actor,omitempty"`
	ActorRole *string   `json:"actor_role,omitempty"`
	ChangedAt time.Time `json:"changed_at"`
	OldData   any       `json:"old_data,omitempty"`
	NewData   any       `json:"new_data,omitempty"`
	Changes   any       `json:"changes,omitempty"`
}

type HistoryFilter struct {
	From   *time.Time
	To     *time.Time
	User   *string
	Action *string
}
