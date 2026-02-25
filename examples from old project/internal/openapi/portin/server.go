package portin

import (
	"context"
	"net/http"

	"gitlab.services.mts.ru/salsa/go-base/application/httphandler/oapi"

	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/converters"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/model"
	"gitlab.services.mts.ru/salsa/mnp-hub/portin-requests/internal/service/portin"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0 -config spec/config.yaml -include-tags portin-orders spec/mnp-portin-service.yaml

const (
	serviceName = "PortIn Orders API"
)

type Server struct {
	portInService portin.IPortInService
}

func NewServer(portInService portin.IPortInService) *Server {
	return &Server{
		portInService: portInService,
	}
}

func (s *Server) GetOApi() (oapi.OpenAPI, error) {
	spec, err := GetSwagger()
	if err != nil {
		return nil, err
	}

	return oapi.NewOpenAPIAdapter(serviceName, spec), nil
}

func (s *Server) NewStrictHandler() http.Handler {
	return Handler(NewStrictHandler(s, nil))
}

func (s *Server) GetPortInOrders(ctx context.Context, req GetPortInOrdersRequestObject) (GetPortInOrdersResponseObject, error) {
	orders, err := s.portInService.SearchPortInOrders(ctx,
		req.Params.Portnumber,
		req.Params.Tempnumber,
		req.Params.CdbProcessId)
	if err != nil {
		return errorResponse(ctx, err), nil
	}

	respList := make([]PortInOrderResponse, len(orders))
	for i, order := range orders {
		respList[i] = mapOrderToOpenAPI(order)
	}

	resp := GetPortInOrders200JSONResponse(respList)

	return &resp, nil
}

func (s *Server) CreatePortInOrder(ctx context.Context, req CreatePortInOrderRequestObject) (CreatePortInOrderResponseObject, error) {
	orderModel := req.Body.ToModel()

	orderID, err := s.portInService.CreatePortInOrder(ctx, orderModel)
	if err != nil {
		return errorResponse(ctx, err), nil
	}

	prefixedID := converters.WithPinPrefix(*orderID)
	resp := CreatePortInOrder200JSONResponse(MnpOrderRef{
		Id:        &prefixedID,
		OrderType: converters.StrToPtr("PortInOrder"),
	})

	return &resp, nil
}

func (s *Server) GetPortInOrder(ctx context.Context, req GetPortInOrderRequestObject) (GetPortInOrderResponseObject, error) {
	if req.Id == "" {
		apiErr := model.ErrMandatoryParameterMissing("id")
		return errorResponse(ctx, apiErr), nil
	}

	order, err := s.portInService.GetPortInOrderByID(ctx, req.Id)
	if err != nil {
		return errorResponse(ctx, err), nil
	}

	apiResp := mapOrderToOpenAPI(order)
	resp := GetPortInOrder200JSONResponse(apiResp)

	return &resp, nil
}
