package io

import coretypes "github.com/drlinggg/gocurl/core/types"

type Output interface {
	Write(req *coretypes.Request, resp *coretypes.Response) error
}
