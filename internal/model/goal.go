package model

import "time"

type Goal struct {
	ID           int64
	Name         string
	TargetAmount float64
	TargetDate   *time.Time
	SoftDelete
}
