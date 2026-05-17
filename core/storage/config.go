package storage

import coretypes "github.com/banakh/gocurl/core/types"

type Config interface {
	FillRequest(req *coretypes.Request) (coretypes.HTTPRequest, error)
	Colors() coretypes.Colors
}
