package mnpevent

import "time"

const (
	ServiceSource           = "mnp-portin-event"
	PortInProcessType       = "portin"
	CreatedEventType        = "created"
	DueDateChangedEventType = "duedate-changed"
)

type PortIn struct {
	ID          string     `json:"id" validate:"required"`
	EventType   string     `json:"eventType" validate:"required"`
	Date        string     `json:"date" validate:"required"`
	ProcessType string     `json:"processType" validate:"required"`
	Source      string     `json:"source" validate:"required"`
	Data        PortInData `json:"data" validate:"required"`
}

type PortInData struct {
	OrderID          string               `json:"orderId,omitempty"`
	OrderState       *OrderStateDTO       `json:"orderState,omitempty"`
	PortationNumbers []PortationNumberDTO `json:"portationNumbers,omitempty"`
	DueDate          *time.Time           `json:"dueDate,omitempty"`
}
