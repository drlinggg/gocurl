package io

import coretypes "github.com/banakh/gocurl/core/types"

type Input interface {
	Read() (*coretypes.Request, error)
}
