package portin

import (
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
)

func mapOrderToOpenAPI(order *model.PortInOrder) PortInOrderResponse {
	prefixedID := converters.WithPinPrefix(order.ID)

	var processType *PortInOrderResponseProcessType

	if order.ProcessType != nil {
		pt := PortInOrderResponseProcessType(*order.ProcessType)
		processType = &pt
	}

	return PortInOrderResponse{
		CdbProcessId:     order.CdbProcessID,
		Comment:          order.Comment,
		Company:          mapCompany(order.Company),
		Contract:         mapMnpDocumentRef(order.Contract),
		Donor:            mapOperator(order.Donor),
		DueDate:          order.DueDate,
		Government:       mapGovernment(order.Government),
		Id:               prefixedID,
		Individual:       mapIndividual(order.Individual),
		Person:           mapPerson(order.Person),
		PortationNumbers: mapPortationNumbers(order.PortationNumbers),
		ProcessType:      processType,
		Recipient:        mapOperator(order.Recipient),
		Source:           order.Source,
		State:            mapOrderState(order.State),
	}
}

func mapOperator(op *model.Operator) *Operator {
	if op == nil {
		return nil
	}

	return &Operator{
		CdbCode:         op.CdbCode,
		Mnc:             op.Mnc,
		Name:            op.Name,
		NetworkOperator: op.NetworkOperator,
		Region:          mapRegion(op.Region),
		Rn:              op.Rn,
	}
}

func mapRegion(r *model.Region) *Region {
	if r == nil {
		return nil
	}

	return &Region{
		Code:  r.Code,
		Kladr: r.Kladr,
		Name:  r.Name,
	}
}

func mapPerson(p *model.Person) *Person {
	if p == nil {
		return nil
	}

	return &Person{
		Customer:      mapPartyRef(p.Customer),
		FirstName:     p.FirstName,
		IdDocuments:   mapIDDocuments(p.IDDocuments),
		LastName:      p.LastName,
		LegalCategory: p.LegalCategory,
		MiddleName:    p.MiddleName,
		Numbers:       mapNumbers(p.Numbers),
	}
}

func mapCompany(c *model.Company) *Company {
	if c == nil {
		return nil
	}

	idDocs := mapIDDocuments(c.IDDocuments)

	return &Company{
		AuthorizedPerson: mapAuthPerson(c.AuthorizedPerson),
		Customer:         mapPartyRef(c.Customer),
		IdDocuments:      &idDocs,
		Inn:              c.Inn,
		Name:             c.Name,
		Numbers:          mapNumbers(c.Numbers),
	}
}

func mapGovernment(g *model.Government) *Government {
	if g == nil {
		return nil
	}

	idDocs := mapIDDocuments(g.IDDocuments)

	return &Government{
		AuthorizedPerson: mapAuthPerson(g.AuthorizedPerson),
		ContractDueDate:  g.ContractDueDate,
		Customer:         mapPartyRef(g.Customer),
		IdDocuments:      &idDocs,
		Inn:              g.Inn,
		Name:             g.Name,
		Numbers:          mapNumbers(g.Numbers),
		TenderId:         g.TenderID,
		TradingFloor:     g.TradingFloor,
	}
}

func mapIndividual(i *model.Individual) *Individual {
	if i == nil {
		return nil
	}

	return &Individual{
		Customer:      mapPartyRef(i.Customer),
		FirstName:     i.FirstName,
		IdDocuments:   mapIDDocuments(i.IDDocuments),
		Inn:           i.Inn,
		LastName:      i.LastName,
		LegalCategory: i.LegalCategory,
		MiddleName:    i.MiddleName,
		Numbers:       mapNumbers(i.Numbers),
	}
}

func mapPortationNumbers(nums []model.PortationNumber) *[]PortationNumber {
	dtoNums := make([]PortationNumber, len(nums))

	for i, n := range nums {
		dtoNums[i] = mapPortationNumber(n)
	}

	return &dtoNums
}

func mapPortationNumber(num model.PortationNumber) PortationNumber {
	var orderState *OrderState

	if num.Status != nil {
		state := mapOrderState(*num.Status)
		orderState = &state
	}

	return PortationNumber{
		Msisdn:       num.Msisdn,
		Status:       orderState,
		TelcoAccount: mapTelcoAccountRef(num.TelcoAccount),
	}
}

func mapTelcoAccountRef(ref model.TelcoAccountRef) TelcoAccountRef {
	return TelcoAccountRef{
		Id:     ref.ID,
		Msisdn: ref.Msisdn,
	}
}

func mapOrderState(st model.OrderState) OrderState {
	return OrderState{
		Code:       st.Code,
		Message:    st.Message,
		Name:       st.Name,
		StatusDate: st.StatusDate,
	}
}

func mapAuthPerson(ap *model.AuthorizedPerson) *AuthorizedPerson {
	if ap == nil {
		return nil
	}

	return &AuthorizedPerson{
		FirstName:  ap.FirstName,
		LastName:   ap.LastName,
		MiddleName: ap.MiddleName,
		Position:   ap.Position,
	}
}

func mapNumbers(nums []string) *[]Phone {
	if len(nums) == 0 {
		return nil
	}

	return &nums
}

func mapIDDocuments(docs []model.IDDocument) []IdDocument {
	dtoDocs := make([]IdDocument, len(docs))

	for i, d := range docs {
		dtoDocs[i] = mapIDDocument(d)
	}

	return dtoDocs
}

func mapIDDocument(idd model.IDDocument) IdDocument {
	return IdDocument{
		DocName:     idd.DocName,
		DocNumber:   idd.DocNumber,
		DocSeries:   idd.DocSeries,
		DocType:     idd.DocType,
		DocumentUrl: idd.DocumentURL,
	}
}

func mapPartyRef(pr *model.PartyRef) *PartyRef {
	if pr == nil {
		return nil
	}

	return &PartyRef{
		Id: pr.ID,
	}
}

func mapMnpDocumentRef(doc model.MnpDocumentRef) *MnpDocumentRef {
	return &MnpDocumentRef{
		Id:           doc.ID,
		DocumentUrl:  doc.DocumentURL,
		DocumentDate: doc.DocumentDate,
	}
}
