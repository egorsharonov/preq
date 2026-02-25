package entities

import (
	"strconv"
	"time"
)

type PortInOrderEntity struct {
	ID            int64      `db:"order_id"`
	CdbProcessID  *int64     `db:"cdb_process_id"`
	CreationDate  time.Time  `db:"creation_date"`
	DueDate       *time.Time `db:"due_date"`
	State         int        `db:"state"`
	OrderType     string     `db:"order_type"`
	CustomerID    string     `db:"customer_id"`
	ContactPhone  string     `db:"contact_phone"`
	OrderData     string     `db:"order_data"` // JSON
	ChangingDate  *time.Time `db:"changing_date"`
	ChangedByUser *string    `db:"changed_by_user"`
	ProcessType   *string    `db:"process_type"`
}

func (e *PortInOrderEntity) StringID() string {
	return strconv.FormatInt(e.ID, 10)
}
