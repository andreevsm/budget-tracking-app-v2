package domain

import (
	"math"
	"strconv"
)

type Money string

func (m Money) Valid() bool {
	if m == "" {
		return false
	}

	n, err := strconv.ParseFloat(string(m), 64)

	if err != nil {
		return false
	}

	return !math.IsNaN(n) && !math.IsInf(n, 0)
}
