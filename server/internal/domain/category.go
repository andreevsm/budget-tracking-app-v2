package domain

import "time"

type Category struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	UserID    string        `json:"user_id"`
	Kind      OperationType `json:"kind"`
	Color     *string       `json:"color"`
	Icon      *string       `json:"icon"`
	ParentID  *string       `json:"parent_id"`
	CreatedAt time.Time     `json:"created_at"`
}
