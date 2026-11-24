package domain

type CreateCategoryInput struct {
	Name     string        `json:"name"`
	Kind     OperationType `json:"kind"`
	Color    *string       `json:"color"`
	Icon     *string       `json:"icon"`
	ParentID *string       `json:"parent_id"`
}
