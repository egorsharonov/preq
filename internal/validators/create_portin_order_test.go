package validators_test

import (
	"context"
	"testing"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/validators"
)

func TestCreatePortInOrderValidatorValidateProcessType(t *testing.T) {
	validProcessTypes := []string{"ShortTimePort", "LongTimePort", "GOS"}
	invalidProcessTypes := []string{"InvalidType", "", "shorttimeport"}

	// Create base valid order with all required fields except ProcessType
	baseOrder := &model.CreatePortinOrder{
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

	// Test valid process types
	for _, processType := range validProcessTypes {
		t.Run("valid process type "+processType, func(t *testing.T) {
			order := *baseOrder // copy base order
			order.ProcessType = &processType
			validator := validators.NewCreatePortInOrderValidator(&order)

			err := validator.Validate(context.Background())
			if err != nil {
				t.Errorf("CreatePortInOrderValidator.Validate() with valid process type %s returned error: %v", processType, err)
			}
		})
	}

	// Test nil process type
	t.Run("nil process type should be valid", func(t *testing.T) {
		order := *baseOrder // copy base order
		order.ProcessType = nil
		validator := validators.NewCreatePortInOrderValidator(&order)

		err := validator.Validate(context.Background())
		if err != nil {
			t.Errorf("CreatePortInOrderValidator.Validate() with nil process type returned error: %v", err)
		}
	})

	// Test invalid process types
	for _, processType := range invalidProcessTypes {
		t.Run("invalid process type "+processType, func(t *testing.T) {
			order := *baseOrder // copy base order
			order.ProcessType = &processType
			validator := validators.NewCreatePortInOrderValidator(&order)

			err := validator.Validate(context.Background())
			if err == nil {
				t.Errorf("CreatePortInOrderValidator.Validate() with invalid process type %s expected error but got nil", processType)
				return
			}

			expectedErrMsg := "processType должно быть одним из: ShortTimePort, LongTimePort, GOS"
			if err.Message != expectedErrMsg {
				t.Errorf("CreatePortInOrderValidator.Validate() error message = %v, want %v", err.Message, expectedErrMsg)
			}
		})
	}
}

func TestCreatePortInOrderValidatorValidateProcessTypeWithValidOrder(t *testing.T) {
	validProcessTypes := []string{"ShortTimePort", "LongTimePort", "GOS"}

	for _, processType := range validProcessTypes {
		t.Run("valid process type "+processType, func(t *testing.T) {
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

			validator := validators.NewCreatePortInOrderValidator(order)

			err := validator.Validate(context.Background())
			if err != nil {
				t.Errorf("CreatePortInOrderValidator.Validate() with valid process type %s returned error: %v", processType, err)
			}
		})
	}
}

func TestCreatePortInOrderValidatorValidateProcessTypeWithInvalidOrder(t *testing.T) {
	invalidProcessType := "InvalidProcessType"
	order := &model.CreatePortinOrder{
		ProcessType: &invalidProcessType,
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

	validator := validators.NewCreatePortInOrderValidator(order)

	err := validator.Validate(context.Background())
	if err == nil {
		t.Error("CreatePortInOrderValidator.Validate() with invalid process type expected error but got nil")
		return
	}

	expectedErrMsg := "processType должно быть одним из: ShortTimePort, LongTimePort, GOS"
	if err.Message != expectedErrMsg {
		t.Errorf("CreatePortInOrderValidator.Validate() error message = %v, want %v", err.Message, expectedErrMsg)
	}
}
