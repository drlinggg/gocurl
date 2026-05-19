package storage

import coretypes "github.com/drlinggg/gocurl/core/types"

type Config interface {
	FillRequest(req *coretypes.Request) (coretypes.HTTPRequest, error)
	Colors() coretypes.Colors
}
