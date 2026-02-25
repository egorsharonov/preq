package containers

import "time"

type OrderPatchUpdate struct {
	OrderID       int64
	State         int
	CdbProcessID  *int64
	DueDate       *time.Time
	OrderDataJSON string
	ChangingDate  *time.Time
	ChangedBy     *string
	UpdateRes     *UpdateResult
}

type UpdateResult struct {
	StatusChanged               bool
	DueDateChanged              bool
	CDBIDChanged                bool
	PortationNumbersStateChaned bool
}

func (r *UpdateResult) AnyChange() bool {
	return r.StatusChanged || r.DueDateChanged || r.CDBIDChanged || r.PortationNumbersStateChaned
}
