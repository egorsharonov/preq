package portin

import (
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func (body *CreatePortInOrderJSONRequestBody) ToModel() *model.CreatePortinOrder {
	var processType *string

	if body.ProcessType != nil {
		pt := string(*body.ProcessType)
		processType = &pt
	}

	return &model.CreatePortinOrder{
		Source:           body.Source,
		DueDate:          &body.DueDate,
		Comment:          body.Comment,
		Donor:            body.Donor.ToModel(),
		Recipient:        body.Recipient.ToModel(),
		Person:           body.Person.ToModel(),
		Company:          body.Company.ToModel(),
		Government:       body.Government.ToModel(),
		Individual:       body.Individual.ToModel(),
		Contract:         body.Contract.ToModel(),
		PortationNumbers: PortationNumbersToModels(body.PortationNumbers),
		ProcessType:      processType,
	}
}

func (o *Operator) ToModel() *model.Operator {
	if o == nil {
		return nil
	}

	return &model.Operator{
		Rn:              o.Rn,
		Mnc:             o.Mnc,
		Name:            o.Name,
		Region:          o.Region.ToModel(),
		NetworkOperator: o.NetworkOperator,
		CdbCode:         o.CdbCode,
	}
}

func (p *Person) ToModel() *model.Person {
	if p == nil {
		return nil
	}

	return &model.Person{
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		MiddleName:    p.MiddleName,
		LegalCategory: p.LegalCategory,
		Customer:      p.Customer.ToModel(),
		IDDocuments:   IDDocsToModel(p.IdDocuments),
		Numbers:       NumbersToModel(p.Numbers),
	}
}

func (c *Company) ToModel() *model.Company {
	if c == nil {
		return nil
	}

	return &model.Company{
		Name:             c.Name,
		Inn:              c.Inn,
		Customer:         c.Customer.ToModel(),
		IDDocuments:      IDDocsToModel(*c.IdDocuments),
		Numbers:          NumbersToModel(c.Numbers),
		AuthorizedPerson: c.AuthorizedPerson.ToModel(),
	}
}

func (g *Government) ToModel() *model.Government {
	if g == nil {
		return nil
	}

	return &model.Government{
		Name:             g.Name,
		Inn:              g.Inn,
		Customer:         g.Customer.ToModel(),
		IDDocuments:      IDDocsToModel(*g.IdDocuments),
		TenderID:         g.TenderId,
		TradingFloor:     g.TradingFloor,
		ContractDueDate:  g.ContractDueDate,
		Numbers:          NumbersToModel(g.Numbers),
		AuthorizedPerson: g.AuthorizedPerson.ToModel(),
	}
}

func (i *Individual) ToModel() *model.Individual {
	if i == nil {
		return nil
	}

	return &model.Individual{
		FirstName:     i.FirstName,
		LastName:      i.LastName,
		MiddleName:    i.MiddleName,
		Inn:           i.Inn,
		LegalCategory: i.LegalCategory,
		Customer:      i.Customer.ToModel(),
		IDDocuments:   IDDocsToModel(i.IdDocuments),
		Numbers:       NumbersToModel(i.Numbers),
	}
}

func (r *Region) ToModel() *model.Region {
	if r == nil {
		return nil
	}

	return &model.Region{
		Code:  r.Code,
		Kladr: r.Kladr,
		Name:  r.Name,
	}
}

func (dr MnpDocumentRef) ToModel() model.MnpDocumentRef {
	return model.MnpDocumentRef{
		ID:           dr.Id,
		DocumentDate: dr.DocumentDate,
		DocumentURL:  dr.DocumentUrl,
	}
}

func (ap *AuthorizedPerson) ToModel() *model.AuthorizedPerson {
	if ap == nil {
		return nil
	}

	return &model.AuthorizedPerson{
		FirstName:  ap.FirstName,
		LastName:   ap.LastName,
		MiddleName: ap.MiddleName,
		Position:   ap.Position,
	}
}

func IDDocsToModel(docs []IdDocument) []model.IDDocument {
	modelDocs := make([]model.IDDocument, len(docs))

	for i, d := range docs {
		modelDocs[i] = d.ToModel()
	}

	return modelDocs
}

func (d IdDocument) ToModel() model.IDDocument {
	return model.IDDocument{
		DocName:     d.DocName,
		DocSeries:   d.DocSeries,
		DocNumber:   d.DocNumber,
		DocumentURL: d.DocumentUrl,
		DocType:     d.DocType,
	}
}

func NumbersToModel(phones *[]Phone) []string {
	if phones == nil {
		return []string{}
	}

	return *phones
}

func PortationNumbersToModels(nums []PortationNumber) []model.PortationNumber {
	modelNums := make([]model.PortationNumber, len(nums))

	for i, n := range nums {
		modelNums[i] = n.ToModel()
	}

	return modelNums
}

func (n PortationNumber) ToModel() model.PortationNumber {
	return model.PortationNumber{
		Msisdn:       n.Msisdn,
		TelcoAccount: n.TelcoAccount.ToModel(),
		Status:       n.Status.ToModel(),
	}
}

func (o *OrderState) ToModel() *model.OrderState {
	if o == nil {
		return nil
	}

	return &model.OrderState{
		Code:       o.Code,
		Message:    o.Message,
		StatusDate: o.StatusDate,
		Name:       o.Name,
	}
}

func (t TelcoAccountRef) ToModel() model.TelcoAccountRef {
	return model.TelcoAccountRef{
		ID:     t.Id,
		Msisdn: t.Msisdn,
	}
}

func (r *PartyRef) ToModel() *model.PartyRef {
	if r == nil {
		return nil
	}

	return &model.PartyRef{
		ID: r.Id,
	}
}
