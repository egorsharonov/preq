package model

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
)

const (
	PortInOrderType    = "portin"
	DefaultProcessType = "ShortTimePort"
)

type CreatePortinOrder struct {
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
	ProcessType      *string
}

func (m *CreatePortinOrder) ToEntity() (*entities.PortInOrderEntity, error) {
	orderDataJSON, err := m.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to convert create order model to db entity: %w", err)
	}

	customerID, err := m.GetCustomerID()
	if err != nil {
		return nil, fmt.Errorf("failed to convert create order model to db entity: %w", err)
	}

	stateCodeName, _ := StateCodeNameToInt("created")

	processType := m.ProcessType
	if processType == nil {
		defaultProcessType := DefaultProcessType
		processType = &defaultProcessType
	}

	entity := &entities.PortInOrderEntity{
		DueDate:       m.DueDate,
		State:         stateCodeName,
		OrderType:     PortInOrderType,
		CustomerID:    customerID,
		ContactPhone:  m.GetContactPhone(),
		OrderData:     orderDataJSON,
		ChangedByUser: nil,
		ProcessType:   processType,
	}

	return entity, nil
}

func (m *CreatePortinOrder) WithDefaults(mtsDefaultOperator *Operator) {
	if m.Contract.DocumentDate == nil {
		now := time.Now().Truncate(24 * time.Hour) // без времени
		m.Contract.DocumentDate = &now
	}

	// Оставить пустым
	// if request.DueDate == nil {
	// dueDate := time.Now().AddDate(0, 0, 8).Truncate(24 * time.Hour) // без времени.
	// request.DueDate = &dueDate
	// }

	if m.Recipient == nil || m.Recipient.Rn == "" ||
		m.Recipient.CdbCode == nil || *m.Recipient.CdbCode == "" {
		m.Recipient = mtsDefaultOperator
	}
}

func (m *CreatePortinOrder) GetCustomerID() (string, error) {
	if err := m.protectFromNil(); err != nil {
		return "", err
	}

	if m.Person != nil && m.Person.Customer != nil {
		return m.Person.Customer.ID, nil
	}

	if m.Company != nil && m.Company.Customer != nil {
		return m.Company.Customer.ID, nil
	}

	if m.Government != nil && m.Government.Customer != nil {
		return m.Government.Customer.ID, nil
	}

	if m.Individual != nil && m.Individual.Customer != nil { // поддержка ИП
		return m.Individual.Customer.ID, nil
	}

	return "", fmt.Errorf("failed to get any customer from create order model")
}

func (m *CreatePortinOrder) GetMSISDNs() []string {
	msisdns := make([]string, len(m.PortationNumbers))
	for i, portNum := range m.PortationNumbers {
		msisdns[i] = portNum.Msisdn
	}

	return msisdns
}

func (m *CreatePortinOrder) GetContactPhone() string {
	if len(m.PortationNumbers) > 0 {
		return m.PortationNumbers[0].Msisdn
	}

	return ""
}

func (m *CreatePortinOrder) ToJSON() (string, error) {
	if err := m.protectFromNil(); err != nil {
		return "", err
	}

	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to serialize create order model: %w", err)
	}

	return string(data), nil
}

func (m *CreatePortinOrder) protectFromNil() error {
	if m == nil {
		return fmt.Errorf("unexpected nil value on create order model")
	}

	return nil
}
