package storage

import coretypes "github.com/drlinggg/gocurl/core/types"

type History interface {
	Save(req *coretypes.Request, resp *coretypes.Response) error
}
