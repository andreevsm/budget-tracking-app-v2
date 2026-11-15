package domain

import "time"

type Date string

func (d Date) Valid() bool {
	if d == "" {
		return false
	}

	_, err := time.Parse("2006-01-02", string(d))

	return err == nil
}
