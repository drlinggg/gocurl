package io

import coretypes "github.com/drlinggg/gocurl/core/types"

type Input interface {
	Read() (*coretypes.Request, error)
}
