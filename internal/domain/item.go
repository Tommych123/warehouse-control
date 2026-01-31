package domain

import "time"

type Item struct {
	ID       int64     `json:"id"`
	SKU      string    `json:"sku"`
	Name     string    `json:"name"`
	Qty      int       `json:"qty"`
	Location *string   `json:"location,omitempty"`
	Created  time.Time `json:"created_at"`
	Updated  time.Time `json:"updated_at"`
}

type ItemCreate struct {
	SKU      string
	Name     string
	Qty      int
	Location *string
}

type ItemUpdate struct {
	SKU      string
	Name     string
	Qty      int
	Location *string
}
