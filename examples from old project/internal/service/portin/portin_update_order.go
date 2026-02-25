package portin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gitlab.services.mts.ru/salsa/go-base/application/diagnostics"
	"go.uber.org/zap"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/containers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/dal/entities"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model/mappers"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/orchestrator/kafka/mnpevent"
)

func (s *Service) UpdatePortInOrder(ctx context.Context, patch *mnpevent.PortInPatch) error {
	tracer := diagnostics.TracerFromContext(ctx)

	ctx, span := tracer.Start(ctx, "PortInService.UpdatePortInOrder")
	defer span.End()

	log := s.getNamedLogger(ctx)

	getBy, err := s.validator.ValidatePortInPatchEvent(ctx, patch)
	if err != nil {
		log.Error("validatePatchEvent.fail", zap.Error(err))
		return err
	}

	getBy.ForUpdate = true

	tx, rollback, err := s.repo.CreateAndOpenTransaction(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}
	defer rollback()

	orderEntity, err := s.getPortInOrder(ctx, tx, getBy)
	if err != nil {
		return err
	}

	if err := s.validator.ValidatePortInPatchDate(patch, orderEntity); err != nil {
		log.Error("validatePortInPatchDate.fail", zap.Error(err))
		return err
	}

	patchRes, order, err := s.formPatchUpdate(orderEntity, patch)
	if err != nil {
		return err
	}

	if patchRes.UpdateRes.AnyChange() {
		err = s.repo.UpdateOrderPatch(ctx, tx, patchRes)
		if err != nil {
			log.Error("PortInOrdersRepository.UpdateOrder.fail", zap.Error(err), zap.Int64("orderId", orderEntity.ID))
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if patchRes.UpdateRes.AnyChange() {
		log.Info("PortInOrdersRepository.UpdateOrder.success",
			zap.Int64("orderId", orderEntity.ID),
			zap.Int64p("cdbProcessId", orderEntity.CdbProcessID),
			zap.Bool("statusChanged", patchRes.UpdateRes.StatusChanged),
			zap.Bool("dueDateChanged", patchRes.UpdateRes.DueDateChanged),
			zap.Bool("cdbIdChanged", patchRes.UpdateRes.CDBIDChanged),
			zap.Bool("portationNumbersStateChanged", patchRes.UpdateRes.PortationNumbersStateChaned),
		)

		if err := s.publishUpdateEvent(ctx, patchRes.UpdateRes, order, patch); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) formPatchUpdate(
	orderEntity *entities.PortInOrderEntity, patch *mnpevent.PortInPatch) (*containers.OrderPatchUpdate, *model.PortInOrder, error) {
	order, err := mappers.PortInEntityToModel(orderEntity)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map order entity to model: %w", err)
	}

	res, err := s.applyPatch(order, patch)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to apply patch data: %w", err)
	}

	orderIntID, err := order.IntID()
	if err != nil {
		return nil, nil, err
	}

	cdbPIDInt, err := order.IntCDBPID()
	if err != nil {
		return nil, nil, err
	}

	updatedOrderJSON, err := order.ToJSON()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse order to JSON: %w", err)
	}

	intStateCode, err := model.StateCodeNameToInt(order.State.Code)
	if err != nil {
		return nil, nil, err
	}

	applyPatch := &containers.OrderPatchUpdate{
		OrderID:       orderIntID,
		State:         intStateCode,
		CdbProcessID:  cdbPIDInt,
		DueDate:       order.DueDate,
		OrderDataJSON: updatedOrderJSON,
		ChangingDate:  &patch.Date,
		ChangedBy:     &patch.Source,
		UpdateRes:     res,
	}

	return applyPatch, order, nil
}

func (s *Service) applyPatch(
	order *model.PortInOrder,
	patch *mnpevent.PortInPatch,
) (*containers.UpdateResult, error) {
	applyState, err := s.applyPatchState(order, patch.Data.OrderState)
	if err != nil {
		return nil, err
	}

	applyDueDate, err := s.applyPatchDueDate(order, patch.Data.DueDate)
	if err != nil {
		return nil, err
	}

	applyCDBPID := s.applyCDBProcessID(order, patch.Data.CDBProcessID)

	numbersMsisdnIdxMap := make(map[string]int, len(order.PortationNumbers))

	for idx, num := range order.PortationNumbers {
		numbersMsisdnIdxMap[num.Msisdn] = idx
	}

	applyPortNums, err := s.applyPortationNumbersState(order, numbersMsisdnIdxMap, patch.Data.PortationNumbers)
	if err != nil {
		return nil, err
	}

	applyAgreagatedState := false
	if applyPortNums {
		applyAgreagatedState = s.applyPortNumbersAggregatedState(order, numbersMsisdnIdxMap, patch.Data.PortationNumbers)
	}

	if applyAgreagatedState {
		applyState = applyAgreagatedState
	}

	return &containers.UpdateResult{
		StatusChanged:               applyState,
		DueDateChanged:              applyDueDate,
		CDBIDChanged:                applyCDBPID,
		PortationNumbersStateChaned: applyPortNums,
	}, nil
}

func (s *Service) applyPatchState(order *model.PortInOrder, patchState *mnpevent.OrderStateDTO) (bool, error) {
	if patchState == nil {
		return false, nil
	}

	code := patchState.Code
	if err := model.ValidCodeName(code); err != nil {
		return false, err
	}

	apply := order.State.Code != code

	order.State.Code = code
	order.State.Message = patchState.Message
	order.State.Name = patchState.Name

	if patchState.StatusDate != nil {
		parsedStatusDate, err := time.Parse(time.RFC3339, *patchState.StatusDate)
		if err != nil {
			return false, model.ErrInvalidDateFormat("orderState.statusDate")
		}

		order.State.StatusDate = &parsedStatusDate
	}

	return apply, nil
}

func (s *Service) applyPatchDueDate(order *model.PortInOrder, patchDueDate string) (bool, error) {
	patchDDT := strings.TrimSpace(patchDueDate)
	if patchDDT == "" {
		return false, nil
	}

	parsed, err := time.Parse(time.RFC3339, patchDDT)
	if err != nil {
		return false, model.ErrInvalidDateFormat("dueDate")
	}

	apply := !order.DueDate.Equal(parsed)
	order.DueDate = &parsed

	return apply, nil
}

func (s *Service) applyCDBProcessID(order *model.PortInOrder, patchCDBPID string) bool {
	if patchCDBPID != "" && order.CdbProcessID == nil {
		order.CdbProcessID = &patchCDBPID
		return true
	}

	return false
}

// Номера уже отвалидирован на этапе валидации.
func (s *Service) applyPortationNumbersState(
	order *model.PortInOrder,
	numbersMsisdnIdxMap map[string]int,
	patchPortationNumbers []mnpevent.PortationNumberDTO) (applyPortNums bool, err error) {
	if len(patchPortationNumbers) == 0 {
		return false, nil
	}

	hasChanged := false

	for patchIdx, patchNum := range patchPortationNumbers {
		idx, ok := numbersMsisdnIdxMap[patchNum.MSISDN]
		// Проверка что patch номер принадлежит заявке и если принадлежит его статус меняется
		if !ok || idx < 0 || idx >= len(order.PortationNumbers) ||
			(order.PortationNumbers[idx].Status != nil &&
				order.PortationNumbers[idx].Status.Code == patchNum.Status.Code) {
			continue
		}

		newStatus := &model.OrderState{
			Code:    patchNum.Status.Code,
			Message: patchNum.Status.Message,
			Name:    patchNum.Status.Name,
		}

		if patchNum.Status.StatusDate != nil {
			parsedStatusDate, err := time.Parse(time.RFC3339, *patchNum.Status.StatusDate)
			if err != nil {
				return false, model.ErrInvalidDateFormat(fmt.Sprintf("portationNumers[%d].status.statusDate", patchIdx))
			}

			newStatus.StatusDate = &parsedStatusDate
		}

		order.PortationNumbers[idx].Status = newStatus
		hasChanged = true
	}

	return hasChanged, nil
}

func (s *Service) applyPortNumbersAggregatedState(
	order *model.PortInOrder,
	numbersMsisdnIdxMap map[string]int,
	patchNums []mnpevent.PortationNumberDTO) bool {
	if len(patchNums) == 0 {
		return false
	}

	// Проверяем что есть patch существуюшего в заявке номера на нужный статус
	stateAggregator, hasAggregatedState := s.checkNumsHaveAggregatedState(order, numbersMsisdnIdxMap, patchNums)

	agregatedStateApplied := false

	if hasAggregatedState && stateAggregator != nil {
		correctStateCount := 0
		// Проверяем что все статусы номеров в заявке лежат в множестве необходимых статусов
		for _, ordNum := range order.PortationNumbers {
			if stateAggregator.TargerNumbersStates.Contains(ordNum.Status.Code) {
				correctStateCount++
			}
		}

		// Если все номера в заявке теперь в нужных статусах устанавлиаем статус всей заявки на агрегированный по статусам номеров
		if correctStateCount == len(order.PortationNumbers) {
			order.State = stateAggregator.AgregatedOrderState
			agregatedStateApplied = true
		}
	}

	return agregatedStateApplied
}

func (s *Service) checkNumsHaveAggregatedState(
	order *model.PortInOrder,
	numbersMsisdnIdxMap map[string]int,
	patchNums []mnpevent.PortationNumberDTO) (*model.PortationNumberStateAgregate, bool) {
	var stateAggregator *model.PortationNumberStateAgregate

	hasAggregatedState := false

	// Проверяем что есть patch существуюшего в заявке номера на нужный статус
	for _, pNum := range patchNums {
		// Проверяем что номер есть в заявке
		idx, ok := numbersMsisdnIdxMap[pNum.MSISDN]
		if !ok || idx < 0 || idx >= len(order.PortationNumbers) {
			continue
		}

		// Проверяем что статус patch номера входит в множество доступных для последующей аггрегации
		if stateAg, ok := s.portInOrdersConfig.PortationNumbersStatesAgregateMap[pNum.Status.Code]; ok {
			hasAggregatedState = ok
			stateAggregator = &stateAg

			break
		}
	}

	return stateAggregator, hasAggregatedState
}

func (s *Service) publishUpdateEvent(
	ctx context.Context,
	patchUpd *containers.UpdateResult,
	order *model.PortInOrder,
	patch *mnpevent.PortInPatch) error {
	log := s.getNamedLogger(ctx)

	if patchUpd.StatusChanged || patchUpd.PortationNumbersStateChaned {
		// При обновлении state заявки через агрегацию state-ов номеров в заявке не передается PortInPatchEvent.data.state,
		// используем обновленный state заявки (либо = PortInPatchEvent.data.state в случае явной передачи, либо = агрегированному статусу)
		var orderStateDTO *mnpevent.OrderStateDTO
		if patchUpd.StatusChanged {
			orderStateDTO = &mnpevent.OrderStateDTO{
				Code:    order.State.Code,
				Message: order.State.Message,
				Name:    order.State.Name,
			}

			if order.State.StatusDate != nil {
				stausDate := order.State.StatusDate.Format(time.RFC3339)
				orderStateDTO.StatusDate = &stausDate
			}
		}

		err := s.mnpEventProducer.PublishStatusChangedEvent(ctx, orderStateDTO, patch, order.ID)
		if err != nil {
			log.Error("publishStatusChangedEvent.fail", zap.Error(err), zap.String("orderId", order.ID))
			return err
		}
	}

	if patchUpd.DueDateChanged && strings.TrimSpace(patch.Data.DueDate) != "" {
		err := s.mnpEventProducer.PublishDueDateChangedEvent(ctx, patch, order.ID)
		if err != nil {
			log.Error("publishDueDateChangedEvent.fail", zap.Error(err), zap.String("orderId", order.ID))
			return err
		}
	}

	return nil
}
