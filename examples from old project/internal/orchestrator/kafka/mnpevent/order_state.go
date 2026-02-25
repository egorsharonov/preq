package mnpevent

type OrderStateDTO struct {
	Code       string  `json:"code" validate:"required"`
	Message    *string `json:"message,omitempty"`
	StatusDate *string `json:"statusDate,omitempty"`
	Name       *string `json:"name,omitempty"`
}
