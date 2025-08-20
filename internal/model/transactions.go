package model

import "time"

type Trasaction struct {
	ID          int64
	UserID      int64
	Amount      int64
	Category    string
	Description string
	Date        time.Time
}