package server

import (
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func newGraphID() openapi_types.UUID {
	return uuid.New()
}

func newRunID() openapi_types.UUID {
	return uuid.New()
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
