package internal_test

import (
	"context"
	"testing"
	"time"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/validators"
)

func TestProcessTypeIntegration(t *testing.T) {
	t.Run("valid process types", testValidProcessTypes)
	t.Run("nil process type", testNilProcessType)
	t.Run("invalid process types", testInvalidProcessTypes)
}

func testValidProcessTypes(t *testing.T) {
	validProcessTypes := []string{"ShortTimePort", "LongTimePort", "GOS"}

	for _, processType := range validProcessTypes {
		t.Run("valid process type "+processType, func(t *testing.T) {
			order := createTestOrder(&processType)
			testOrderProcessType(t, order, processType, false)
		})
	}
}

func testNilProcessType(t *testing.T) {
	t.Run("nil process type should use default", func(t *testing.T) {
		order := createTestOrder(nil)
		testOrderProcessType(t, order, model.DefaultProcessType, false)
	})
}

func testInvalidProcessTypes(t *testing.T) {
	invalidProcessTypes := []string{"InvalidType", "WrongType"}

	for _, processType := range invalidProcessTypes {
		t.Run("invalid process type "+processType, func(t *testing.T) {
			order := createTestOrder(&processType)
			testOrderProcessType(t, order, "", true)
		})
	}
}

func createTestOrder(processType *string) *model.CreatePortinOrder {
	return &model.CreatePortinOrder{
		ProcessType: processType,
		Person: &model.Person{
			FirstName: "Test",
			LastName:  "User",
			Customer: &model.PartyRef{
				ID: "123",
			},
			IDDocuments: []model.IDDocument{
				{
					DocNumber: "123456",
					DocType:   converters.ToPtr("CPassport"),
				},
			},
		},
		Contract: model.MnpDocumentRef{
			DocumentURL: "http://example.com/doc.pdf",
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("123"),
				},
			},
		},
	}
}

func testOrderProcessType(t *testing.T, order *model.CreatePortinOrder, expectedType string, shouldError bool) {
	ctx := context.Background()

	// Test validation
	validator := validators.NewCreatePortInOrderValidator(order)
	err := validator.Validate(ctx)

	if shouldError {
		if err == nil {
			t.Errorf("Expected validation error for processType %v, but got nil", order.ProcessType)
		}

		return
	}

	if err != nil {
		t.Errorf("Unexpected validation error for processType %v: %v", order.ProcessType, err)
		return
	}

	// Test conversion to entity
	entity, convertErr := order.ToEntity()
	if convertErr != nil {
		t.Fatalf("Failed to convert order to entity: %v", convertErr)
	}

	if entity.ProcessType == nil {
		t.Errorf("Entity ProcessType is nil, want %s", expectedType)
		return
	}

	if *entity.ProcessType != expectedType {
		t.Errorf("Entity ProcessType = %v, want %v", *entity.ProcessType, expectedType)
	}

	// Test JSON serialization
	jsonStr, jsonErr := order.ToJSON()
	if jsonErr != nil {
		t.Fatalf("Failed to serialize order to JSON: %v", jsonErr)
	}

	// Verify that JSON contains the processType (if not nil)
	if order.ProcessType != nil && !containsSubstring(jsonStr, *order.ProcessType) {
		t.Errorf("JSON does not contain processType %s. JSON: %s", *order.ProcessType, jsonStr)
	}
}

func TestProcessTypeCompleteOrderExample(t *testing.T) {
	ctx := context.Background()

	// Test with the example from the task description
	processType := "GOS"
	dueDate := time.Date(2026, 3, 2, 11, 7, 54, 614813100, time.UTC)
	comment := "Test portin request for creation person"
	source := "MNPHub.E2E.Testing.Local"

	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		DueDate:     &dueDate,
		Comment:     &comment,
		Source:      &source,
		Person: &model.Person{
			FirstName:     "Кролик",
			LastName:      "Белый",
			MiddleName:    converters.ToPtr("Маркович"),
			LegalCategory: converters.ToPtr("резидент"),
			Customer: &model.PartyRef{
				ID: "string",
			},
			IDDocuments: []model.IDDocument{
				{
					DocName:     converters.ToPtr("Паспорт гражданина РФ"),
					DocNumber:   "678901",
					DocSeries:   converters.ToPtr("9876"),
					DocType:     converters.ToPtr("CPassport"),
					DocumentURL: converters.ToPtr("CHANGES/mnphub/1-633840142983/79160141384.PORTIN.pdf"),
				},
			},
			Numbers: []string{"79167498100"},
		},
		Contract: model.MnpDocumentRef{
			ID:           converters.ToPtr("1234567"),
			DocumentURL:  "CHANGES/mnphub/1-633840142983/79160141384.PORTIN.pdf",
			DocumentDate: &time.Time{},
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79167498110",
				TelcoAccount: model.TelcoAccountRef{
					ID:     converters.ToPtr("98765432"),
					Msisdn: converters.ToPtr("79163803292"),
				},
			},
		},
	}

	// Test validation
	validator := validators.NewCreatePortInOrderValidator(order)

	err := validator.Validate(ctx)
	if err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Test conversion to entity
	entity, convertErr := order.ToEntity()
	if convertErr != nil {
		t.Fatalf("Failed to convert order to entity: %v", convertErr)
	}

	// Verify ProcessType
	if entity.ProcessType == nil {
		t.Error("Entity ProcessType is nil")
		return
	}

	if *entity.ProcessType != processType {
		t.Errorf("Entity ProcessType = %v, want %v", *entity.ProcessType, processType)
	}

	// Verify other fields
	if entity.OrderType != model.PortInOrderType {
		t.Errorf("Entity OrderType = %v, want %v", entity.OrderType, model.PortInOrderType)
	}

	if entity.CustomerID != "string" {
		t.Errorf("Entity CustomerID = %v, want %v", entity.CustomerID, "string")
	}

	if entity.ContactPhone != "79167498110" {
		t.Errorf("Entity ContactPhone = %v, want %v", entity.ContactPhone, "79167498110")
	}

	// Test JSON serialization
	jsonStr, jsonErr := order.ToJSON()
	if jsonErr != nil {
		t.Fatalf("Failed to serialize order to JSON: %v", jsonErr)
	}

	// Verify that JSON contains the processType
	if !containsSubstring(jsonStr, processType) {
		t.Errorf("JSON does not contain processType %s. JSON: %s", processType, jsonStr)
	}

	// Verify that JSON contains other expected fields
	expectedFields := []string{
		"Кролик",
		"Белый",
		"79167498110",
		"Test portin request for creation person",
	}

	for _, field := range expectedFields {
		if !containsSubstring(jsonStr, field) {
			t.Errorf("JSON does not contain expected field %s. JSON: %s", field, jsonStr)
		}
	}
}

func TestProcessTypeWithDefaults(t *testing.T) {
	processType := "LongTimePort"
	order := &model.CreatePortinOrder{
		ProcessType: &processType,
		Person: &model.Person{
			FirstName: "Test",
			LastName:  "User",
			Customer: &model.PartyRef{
				ID: "123",
			},
			IDDocuments: []model.IDDocument{
				{
					DocNumber: "123456",
					DocType:   converters.ToPtr("CPassport"),
				},
			},
		},
		Contract: model.MnpDocumentRef{
			DocumentURL: "http://example.com/doc.pdf",
		},
		PortationNumbers: []model.PortationNumber{
			{
				Msisdn: "79123456789",
				TelcoAccount: model.TelcoAccountRef{
					ID: converters.ToPtr("123"),
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
		t.Error("ProcessType became nil after WithDefaults")
		return
	}

	if *order.ProcessType != processType {
		t.Errorf("ProcessType = %v, want %v", *order.ProcessType, processType)
	}

	// Verify that DocumentDate was set
	if order.Contract.DocumentDate == nil {
		t.Error("DocumentDate was not set by WithDefaults")
	}

	// Verify that Recipient was set to default
	if order.Recipient == nil {
		t.Error("Recipient was not set by WithDefaults")
	} else if order.Recipient.Rn != mtsDefaultOperator.Rn {
		t.Errorf("Recipient.Rn = %v, want %v", order.Recipient.Rn, mtsDefaultOperator.Rn)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && containsSubstring(s[1:], substr)) ||
		containsSubstring(s, substr[1:]))
}
