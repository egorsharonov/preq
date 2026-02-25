package model_test

import (
	"testing"
	"time"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func TestCreatePortinOrderToEntityWithProcessType(t *testing.T) {
	tests := []struct {
		name         string
		processType  *string
		expectedType string
	}{
		{
			name:         "nil process type should use default",
			processType:  nil,
			expectedType: model.DefaultProcessType,
		},
		{
			name:         "ShortTimePort should be preserved",
			processType:  converters.ToPtr("ShortTimePort"),
			expectedType: "ShortTimePort",
		},
		{
			name:         "LongTimePort should be preserved",
			processType:  converters.ToPtr("LongTimePort"),
			expectedType: "LongTimePort",
		},
		{
			name:         "GOS should be preserved",
			processType:  converters.ToPtr("GOS"),
			expectedType: "GOS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &model.CreatePortinOrder{
				ProcessType: tt.processType,
				Person: &model.Person{
					FirstName: "Test",
					LastName:  "User",
					Customer: &model.PartyRef{
						ID: "123",
					},
				},
				Contract: model.MnpDocumentRef{
					DocumentURL: "http://example.com/doc.pdf",
				},
				PortationNumbers: []model.PortationNumber{
					{
						Msisdn: "79123456789",
						TelcoAccount: model.TelcoAccountRef{
							ID: converters.ToPtr("12345"),
						},
					},
				},
			}

			entity, err := order.ToEntity()
			if err != nil {
				t.Fatalf("CreatePortinOrder.ToEntity() error = %v", err)
			}

			if entity.ProcessType == nil {
				t.Errorf("CreatePortinOrder.ToEntity() ProcessType is nil, want %s", tt.expectedType)
				return
			}

			if *entity.ProcessType != tt.expectedType {
				t.Errorf("CreatePortinOrder.ToEntity() ProcessType = %v, want %v", *entity.ProcessType, tt.expectedType)
			}
		})
	}
}

func TestCreatePortinOrderToEntityWithValidOrder(t *testing.T) {
	processType := "GOS"
	dueDate := time.Now().AddDate(0, 0, 8)
	comment := "Test order"
	source := "test"

	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		DueDate:     &dueDate,
		Comment:     &comment,
		Source:      &source,
		Person: &model.Person{
			FirstName: "Иван",
			LastName:  "Иванов",
			Customer: &model.PartyRef{
				ID: "customer123",
			},
		},
		Contract: model.MnpDocumentRef{
			DocumentURL: "http://example.com/contract.pdf",
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("account123"),
				},
			},
		},
	}

	entity, err := order.ToEntity()
	if err != nil {
		t.Fatalf("CreatePortinOrder.ToEntity() error = %v", err)
	}

	// Verify basic fields
	if entity.OrderType != model.PortInOrderType {
		t.Errorf("CreatePortinOrder.ToEntity() OrderType = %v, want %v", entity.OrderType, model.PortInOrderType)
	}

	if entity.CustomerID != "customer123" {
		t.Errorf("CreatePortinOrder.ToEntity() CustomerID = %v, want %v", entity.CustomerID, "customer123")
	}

	if entity.ContactPhone != "79123456789" {
		t.Errorf("CreatePortinOrder.ToEntity() ContactPhone = %v, want %v", entity.ContactPhone, "79123456789")
	}

	// Verify ProcessType
	if entity.ProcessType == nil {
		t.Error("CreatePortinOrder.ToEntity() ProcessType is nil")
		return
	}

	if *entity.ProcessType != processType {
		t.Errorf("CreatePortinOrder.ToEntity() ProcessType = %v, want %v", *entity.ProcessType, processType)
	}
}

func TestCreatePortinOrderToJSONWithProcessType(t *testing.T) {
	processType := "LongTimePort"
	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		Person: &model.Person{
			FirstName: "Test",
			LastName:  "User",
			Customer: &model.PartyRef{
				ID: "123",
			},
		},
		Contract: model.MnpDocumentRef{
			DocumentURL: "http://example.com/doc.pdf",
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("12345"),
				},
			},
		},
	}

	jsonStr, err := order.ToJSON()
	if err != nil {
		t.Fatalf("CreatePortinOrder.ToJSON() error = %v", err)
	}

	// Verify that JSON contains the processType
	if !containsSubstring(jsonStr, processType) {
		t.Errorf("CreatePortinOrder.ToJSON() JSON does not contain processType %s. JSON: %s", processType, jsonStr)
	}
}

func TestCreatePortinOrderWithDefaults(t *testing.T) {
	processType := "GOS"
	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		Person: &model.Person{
			FirstName: "Test",
			LastName:  "User",
			Customer: &model.PartyRef{
				ID: "123",
			},
		},
		Contract: model.MnpDocumentRef{
			DocumentURL: "http://example.com/doc.pdf",
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("12345"),
				},
			},
		},
	}

	mtsDefaultOperator := &model.Operator{
		Rn:      "D7701",
		CdbCode: converters.ToPtr("mMTS5"),
		Name:    converters.ToPtr("PAO MTS"),
		Mnc:     converters.ToPtr("1"),
		Region: &model.Region{
			Code:  "77",
			Kladr: converters.ToPtr("77"),
			Name:  converters.ToPtr("Moscow"),
		},
	}

	order.WithDefaults(mtsDefaultOperator)

	// Verify that ProcessType is not changed by WithDefaults
	if order.ProcessType == nil {
		t.Error("CreatePortinOrder.WithDefaults() ProcessType became nil")
		return
	}

	if *order.ProcessType != processType {
		t.Errorf("CreatePortinOrder.WithDefaults() ProcessType = %v, want %v", *order.ProcessType, processType)
	}

	// Verify that DocumentDate was set
	if order.Contract.DocumentDate == nil {
		t.Error("CreatePortinOrder.WithDefaults() DocumentDate was not set")
	}

	// Verify that Recipient was set to default
	if order.Recipient == nil {
		t.Error("CreatePortinOrder.WithDefaults() Recipient was not set")
	} else if order.Recipient.Rn != mtsDefaultOperator.Rn {
		t.Errorf("CreatePortinOrder.WithDefaults() Recipient.Rn = %v, want %v", order.Recipient.Rn, mtsDefaultOperator.Rn)
	}
}

func TestCreatePortinOrderGetMSISDNs(t *testing.T) {
	processType := "ShortTimePort"
	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		PortationNumbers: []model.PortationNumber{
			{Msisdn: "79123456789"},
			{Msisdn: "79123456790"},
		},
	}

	msisdns := order.GetMSISDNs()
	if len(msisdns) != 2 {
		t.Errorf("CreatePortinOrder.GetMSISDNs() returned %d numbers, want 2", len(msisdns))
	}

	if msisdns[0] != "79123456789" {
		t.Errorf("CreatePortinOrder.GetMSISDNs()[0] = %v, want %v", msisdns[0], "79123456789")
	}

	if msisdns[1] != "79123456790" {
		t.Errorf("CreatePortinOrder.GetMSISDNs()[1] = %v, want %v", msisdns[1], "79123456790")
	}
}

func TestCreatePortinOrderGetContactPhone(t *testing.T) {
	processType := "LongTimePort"
	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("12345"),
				},
			},
		},
	}

	contactPhone := order.GetContactPhone()
	if contactPhone != "79123456789" {
		t.Errorf("CreatePortinOrder.GetContactPhone() = %v, want %v", contactPhone, "79123456789")
	}

	// Test with empty PortationNumbers
	order.PortationNumbers = []model.PortationNumber{}

	contactPhone = order.GetContactPhone()
	if contactPhone != "" {
		t.Errorf("CreatePortinOrder.GetContactPhone() with empty numbers = %v, want empty string", contactPhone)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && containsSubstring(s[1:], substr)) ||
		containsSubstring(s, substr[1:]))
}
