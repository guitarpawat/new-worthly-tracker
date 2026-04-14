package model

import "time"

type SoftDelete struct {
	DeletedAt *time.Time
}
