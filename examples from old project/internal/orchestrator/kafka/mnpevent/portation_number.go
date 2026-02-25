package mnpevent

type PortationNumberDTO struct {
	MSISDN       string             `json:"msisdn" validate:"required"`
	TelcoAccount TelcoAccountRefDTO `json:"telcoAccount" validate:"required"`
	Status       *OrderStateDTO     `json:"status,omitempty"`
}

type TelcoAccountRefDTO struct {
	ID     *string `json:"id,omitempty"`
	MSISDN *string `json:"msisdn,omitempty"`
}
