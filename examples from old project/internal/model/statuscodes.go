package model

import (
	"fmt"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/datastructures"
)

const (
	portationCompleteStateCode = "portation-complete"
	transferedStateCode        = "transfered" //nolint:misspell // "transfered" используется намеренно для соответствия бизнес-логике
)

type PortationNumberStateAgregate struct {
	// множество статусов, в котором должны лежать все номера в заявке,
	// для применения указанного статуса ко всей заявке
	TargerNumbersStates datastructures.HashSet[string]
	// соотвествующий агренированный статус заявки
	AgregatedOrderState OrderState
}

func GetDefaultPortationNumbersStatesAgregateMap() map[string]PortationNumberStateAgregate {
	return map[string]PortationNumberStateAgregate{
		portationCompleteStateCode: {
			TargerNumbersStates: datastructures.HashSet[string]{
				portationCompleteStateCode: {},
				transferedStateCode:        {},
			},
			AgregatedOrderState: OrderState{
				Code: portationCompleteStateCode,
				Name: converters.ToPtr("Номер перенесен"),
			},
		},
		transferedStateCode: {
			TargerNumbersStates: datastructures.HashSet[string]{
				transferedStateCode: {},
			},
			AgregatedOrderState: OrderState{
				Code: "closed",
				Name: converters.ToPtr("Перенос успешно завершен"),
			},
		},
	}
}

func StateCodeNamesToInts(stateNames []string) ([]int, error) {
	allowedStatuses := make([]int, 0)

	for _, statusName := range stateNames {
		statusCode, err := StateCodeNameToInt(statusName)
		if err != nil {
			return allowedStatuses, err
		}

		allowedStatuses = append(allowedStatuses, statusCode)
	}

	return allowedStatuses, nil
}

func StateIntCodeToName(code int) (string, error) {
	if v, ok := statusCodeToName[code]; ok {
		return v, nil
	}

	return "unknown", fmt.Errorf("unknown state code=%d", code)
}

func StateCodeNameToInt(name string) (int, error) {
	if code, ok := statusNameToCode[name]; ok {
		return code, nil
	}

	return 0, fmt.Errorf("unknown state name=%s", name)
}

func ValidCodeName(name string) error {
	if _, ok := statusNameToCode[name]; ok {
		return nil
	}

	return fmt.Errorf("unknown state name=%s", name)
}

func GetDefaultAllowedStatusesForNewOrder() []int {
	allowedCodesCopy := make([]int, len(defaultAllowedStatusesForNewOrder))
	allowedCodesCopy = append(allowedCodesCopy[:0], defaultAllowedStatusesForNewOrder...)

	return allowedCodesCopy
}

var defaultAllowedStatusesForNewOrder = []int{-1, -2, -3, -4, 4, 6, 22}

var statusNameToCode = map[string]int{
	"cancel-rejected":          -51,
	"arbitation-timeout":       -4,
	"donor-rejected":           -3,
	"canceled":                 -2,
	"cdb-rejected":             -1,
	"created":                  0,
	"sent-cdb":                 1,
	"arbitration":              2,
	"donor-verification":       3,
	"arbitation-pending":       4,
	"debt-checking":            5,
	"debt-collection":          6,
	"portation-waitng":         7,
	"portation-due":            8,
	"portation-ready":          9,
	"portation-exec":           20,
	portationCompleteStateCode: 21,
	"closed":                   22,
	"duedate-changed":          23,
	"cancel-request":           50,
	"cancel-confirmed":         51,
}

var statusCodeToName = map[int]string{
	-51: "cancel-rejected",
	-4:  "arbitation-timeout",
	-3:  "donor-rejected",
	-2:  "canceled",
	-1:  "cdb-rejected",
	0:   "created",
	1:   "sent-cdb",
	2:   "arbitration",
	3:   "donor-verification",
	4:   "arbitation-pending",
	5:   "debt-checking",
	6:   "debt-collection",
	7:   "portation-waitng",
	8:   "portation-due",
	9:   "portation-ready",
	20:  "portation-exec",
	21:  portationCompleteStateCode,
	22:  "closed",
	23:  "duedate-changed",
	50:  "cancel-request",
	51:  "cancel-confirmed",
}
