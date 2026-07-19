package server

import (
	"context"

	"github.com/madmmas/temflowral/backend/internal/api"
)

// API implements the generated strict server interface. Endpoint behavior is
// added here as backend features land; keeping this separate from generated
// code ensures regeneration cannot overwrite application logic.
type API struct{}

var _ api.StrictServerInterface = (*API)(nil)

// NewAPI returns the implementation of the generated HTTP API contract.
func NewAPI() *API {
	return &API{}
}

func (*API) CreateGraph(
	_ context.Context,
	_ api.CreateGraphRequestObject,
) (api.CreateGraphResponseObject, error) {
	return api.CreateGraph500JSONResponse{InternalErrorJSONResponse: notImplementedError()}, nil
}

func (*API) GetGraph(
	_ context.Context,
	_ api.GetGraphRequestObject,
) (api.GetGraphResponseObject, error) {
	return api.GetGraph500JSONResponse{InternalErrorJSONResponse: notImplementedError()}, nil
}

func (*API) StartGraphRun(
	_ context.Context,
	_ api.StartGraphRunRequestObject,
) (api.StartGraphRunResponseObject, error) {
	return api.StartGraphRun500JSONResponse{InternalErrorJSONResponse: notImplementedError()}, nil
}

func (*API) ListNodeTypes(
	_ context.Context,
	_ api.ListNodeTypesRequestObject,
) (api.ListNodeTypesResponseObject, error) {
	return api.ListNodeTypes500JSONResponse{InternalErrorJSONResponse: notImplementedError()}, nil
}

func (*API) GetRun(
	_ context.Context,
	_ api.GetRunRequestObject,
) (api.GetRunResponseObject, error) {
	return api.GetRun500JSONResponse{InternalErrorJSONResponse: notImplementedError()}, nil
}

func notImplementedError() api.InternalErrorJSONResponse {
	code := "not_implemented"
	return api.InternalErrorJSONResponse{
		Code:    &code,
		Message: "endpoint implementation is not available yet",
	}
}
