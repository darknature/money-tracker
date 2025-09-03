package model

import "time"

type Trasaction struct {
	UserID      int64		`json:"-"`
	Amount      int64		`json:"amount"`
	Category    string		`json:"category"`
	Description string		`json:"description"`
	Date        time.Time	`json:"-"`
}