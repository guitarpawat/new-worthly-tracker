package model

import "time"

type RecordSnapshot struct {
	ID         int64
	RecordDate time.Time
	SoftDelete
}
