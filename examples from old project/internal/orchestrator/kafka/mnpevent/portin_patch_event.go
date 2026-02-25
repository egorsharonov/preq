package mnpevent

import "time"

type PortInPatch struct {
	ID        string          `json:"id" validate:"required"`
	EventType string          `json:"eventType" validate:"required"`
	Date      time.Time       `json:"date" validate:"required"`
	Source    string          `json:"source" validate:"required"`
	Data      PortinPatchData `json:"data" validate:"required"`
}

type PortinPatchData struct {
	OrderID          string               `json:"orderId"`
	CDBProcessID     string               `json:"cdbProcessId,omitempty"`
	DueDate          string               `json:"dueDate,omitempty"`
	OrderState       *OrderStateDTO       `json:"orderState,omitempty"`
	PortationNumbers []PortationNumberDTO `json:"portationNumbers,omitempty"`
}
