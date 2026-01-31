package http

import (
	"encoding/json"
	"strconv"
)

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}

func compactJSON(v any) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
