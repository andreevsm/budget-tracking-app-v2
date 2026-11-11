package domain

import "github.com/google/uuid"

type CreateCategoryInput struct {
	Name     string       `json:"name"`
	Kind     CategoryKind `json:"kind"`
	Color    *string      `json:"color"`
	Icon     *string      `json:"icon"`
	ParentID *string      `json:"parent_id"`
	UserID   uuid.UUID    `json:"user_id"`
}
