package domain

type Currency struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	MinorUnits int16  `json:"minor_units"`
}
