package model

import "time"

type PortationNumber struct {
	Msisdn       string          `json:"msisdn" validate:"required"`
	TelcoAccount TelcoAccountRef `json:"telcoAccount" validate:"required"`
	Status       *OrderState     `json:"status,omitempty"`
}

type TelcoAccountRef struct {
	ID     *string `json:"id,omitempty"`
	Msisdn *string `json:"msisdn,omitempty"`
}

type MnpDocumentRef struct {
	ID           *string    `json:"id,omitempty"`
	DocumentDate *time.Time `json:"documentDate,omitempty"`
	DocumentURL  string     `json:"documentUrl" validate:"required"`
}

type Contract = MnpDocumentRef

type Person struct {
	FirstName     string       `json:"firstName" validate:"required"`
	LastName      string       `json:"lastName" validate:"required"`
	MiddleName    *string      `json:"middleName,omitempty"`
	LegalCategory *string      `json:"legalCategory,omitempty"`
	Customer      *PartyRef    `json:"customer,omitempty"`
	IDDocuments   []IDDocument `json:"idDocuments" validate:"required,min=1"`
	Numbers       []string     `json:"numbers,omitempty"`
}

type AuthorizedPerson struct {
	FirstName  string  `json:"firstName" validate:"required"`
	LastName   string  `json:"lastName" validate:"required"`
	MiddleName *string `json:"middleName,omitempty"`
	Position   *string `json:"position,omitempty"`
}

type Company struct {
	Name             string            `json:"name" validate:"required"`
	Inn              string            `json:"inn" validate:"required"`
	Customer         *PartyRef         `json:"customer,omitempty"`
	IDDocuments      []IDDocument      `json:"idDocuments,omitempty"`
	Numbers          []string          `json:"numbers,omitempty"`
	AuthorizedPerson *AuthorizedPerson `json:"authorizedPerson,omitempty"`
}
type Government struct {
	Name             string            `json:"name" validate:"required"`
	Inn              string            `json:"inn" validate:"required"`
	Customer         *PartyRef         `json:"customer,omitempty"`
	IDDocuments      []IDDocument      `json:"idDocuments,omitempty"`
	TenderID         *string           `json:"tenderId,omitempty"`
	TradingFloor     *string           `json:"tradingFloor,omitempty"`
	ContractDueDate  *time.Time        `json:"contractDueDate,omitempty"`
	Numbers          []string          `json:"numbers,omitempty"`
	AuthorizedPerson *AuthorizedPerson `json:"authorizedPerson,omitempty"`
}

type Individual struct {
	FirstName     *string      `json:"firstName,omitempty"`
	LastName      *string      `json:"lastName,omitempty"`
	MiddleName    *string      `json:"middleName,omitempty"`
	Inn           *string      `json:"inn,omitempty"`
	LegalCategory *string      `json:"legalCategory,omitempty"`
	Customer      *PartyRef    `json:"customer,omitempty"`
	IDDocuments   []IDDocument `json:"idDocuments" validate:"required,min=1"`
	Numbers       []string     `json:"numbers,omitempty"`
}

type PartyRef struct {
	ID string `json:"id" validate:"required"`
}

type IDDocument struct {
	DocName     *string `json:"docName,omitempty"`
	DocSeries   *string `json:"docSeries,omitempty"`
	DocNumber   string  `json:"docNumber" validate:"required"`
	DocumentURL *string `json:"documentUrl,omitempty"`
	// Для совместимости с текущим API
	DocType *string `json:"docType,omitempty"`
}

type Operator struct {
	Rn              string           `json:"rn" validate:"required"`
	Mnc             *string          `json:"mnc,omitempty"`
	Name            *string          `json:"name,omitempty"`
	Region          *Region          `json:"region,omitempty"`
	NetworkOperator *NetworkOperator `json:"networkOperator,omitempty"`
	CdbCode         *string          `json:"cdbCode,omitempty"`
}

type Region struct {
	Code  string  `json:"code" validate:"required"`
	Kladr *string `json:"kladr,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// Ссылка на Network Operator в Telco.ROI.
type NetworkOperator = string

type OrderState struct {
	Code       string     `json:"code" validate:"required"`
	Message    *string    `json:"message,omitempty"`
	StatusDate *time.Time `json:"statusDate,omitempty"`
	Name       *string    `json:"name,omitempty"`
}
