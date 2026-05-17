package io

import coretypes "github.com/banakh/gocurl/core/types"

type Output interface {
	Write(*coretypes.Response) error
}
