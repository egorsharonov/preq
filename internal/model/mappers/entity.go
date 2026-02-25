package mappers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func PortInEntityToModel(orderEntity *entities.PortInOrderEntity) (*model.PortInOrder, error) {
	var portInData model.PortInOrder
	if err := json.Unmarshal([]byte(orderEntity.OrderData), &portInData); err != nil {
		return nil, fmt.Errorf("failed to deserialaze order data: %w", err)
	}

	stateCodeName, err := model.StateIntCodeToName(orderEntity.State)
	if err != nil {
		return nil, err
	}

	var cdbProcessID *string

	if orderEntity.CdbProcessID != nil {
		formatted := strconv.FormatInt(*orderEntity.CdbProcessID, 10)
		cdbProcessID = &formatted
	}

	order := &model.PortInOrder{
		ID:               strconv.FormatInt(orderEntity.ID, 10),
		CdbProcessID:     cdbProcessID,
		Source:           portInData.Source,
		DueDate:          orderEntity.DueDate,
		Comment:          portInData.Comment,
		PortationNumbers: portInData.PortationNumbers,
		Donor:            portInData.Donor,
		Recipient:        portInData.Recipient,
		Person:           portInData.Person,
		Company:          portInData.Company,
		Government:       portInData.Government,
		Individual:       portInData.Individual,
		Contract:         portInData.Contract,
		State:            portInData.State,
		ProcessType:      orderEntity.ProcessType,
	}

	order.State.Code = stateCodeName

	return order, nil
}
