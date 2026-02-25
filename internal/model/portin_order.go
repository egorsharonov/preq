package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PortInOrder struct {
	ID               string
	CdbProcessID     *string
	Source           *string
	DueDate          *time.Time
	Comment          *string
	Donor            *Operator
	Recipient        *Operator
	Person           *Person
	Company          *Company
	Government       *Government
	Individual       *Individual
	Contract         MnpDocumentRef
	PortationNumbers []PortationNumber
	State            OrderState
	ProcessType      *string
}

func (o *PortInOrder) IntID() (int64, error) {
	i, err := strconv.ParseInt(o.ID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid order id value: %w", err)
	}

	return i, nil
}

func (o *PortInOrder) IntCDBPID() (*int64, error) {
	if o.CdbProcessID == nil {
		return nil, nil
	}

	i, err := strconv.ParseInt(strings.TrimSpace(*o.CdbProcessID), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cdbProcessId value: %w", err)
	}

	return &i, nil
}

func (o *PortInOrder) ToJSON() (string, error) {
	if o == nil {
		return "", fmt.Errorf("unexpected nil value on PortInOrder model")
	}

	data, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("failed to serialize PortInOrder model: %w", err)
	}

	return string(data), nil
}
