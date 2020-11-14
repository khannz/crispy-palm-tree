package models

import "time"

type Order struct {
	OrderID          string
	OrderTypeID      int
	CreatedAt        time.Time
	OrderSource      string
	OrderRawJSON     string
	ServiceManagerID string
}
