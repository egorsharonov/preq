package transform

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type OrderPayload struct {
	Contract struct {
		DocumentDate string `json:"documentDate"`
	} `json:"contract"`
	ProcessType string `json:"processType"`
	Status      struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
	Recipient struct {
		CDBCode string `json:"cdbCode"`
	} `json:"recipient"`
	PortationNumbers []struct {
		MSISDN string `json:"msisdn"`
		RN     string `json:"rn"`
	} `json:"portationNumbers"`
	Person     any `json:"person"`
	Individual any `json:"individual"`
	Company    any `json:"company"`
	Government any `json:"government"`
}

func ParseOrderPayload(raw []byte) (OrderPayload, error) {
	var p OrderPayload
	err := json.Unmarshal(raw, &p)

	return p, err
}

func SubscriberType(p OrderPayload) string {
	switch {
	case p.Person != nil:
		return "Person"
	case p.Individual != nil:
		return "Entrepreneur"
	case p.Company != nil || p.Government != nil:
		return "Org"
	default:
		return ""
	}
}

func ParseContractDate(v string) *time.Time {
	if strings.TrimSpace(v) == "" {
		return nil
	}

	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		ts, err := time.Parse(layout, v)
		if err == nil {
			return &ts
		}
	}

	return nil
}

func ParseRejectReason(state int, msg string) *int {
	if state >= 0 {
		return nil
	}

	left := msg
	if idx := strings.IndexAny(msg, ".,"); idx >= 0 {
		left = msg[:idx]
	}
	left = strings.TrimSpace(left)
	if left == "" {
		return nil
	}

	parsed, err := strconv.Atoi(left)
	if err != nil {
		return nil
	}

	return &parsed
}
