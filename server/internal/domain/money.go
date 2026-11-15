package domain

import (
	"math"
	"strconv"
	"strings"
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

func (m Money) Negative() bool {
	return strings.HasPrefix(string(m), "-")
}

func (n Money) Sign() int {
	f, _ := strconv.ParseFloat(string(n), 64)
	switch {
	case f > 0:
		return 1
	case f < 0:
		return -1
	default:
		return 0
	}
}
