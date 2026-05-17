package storage

import coretypes "github.com/banakh/gocurl/core/types"

type History interface {
	Save(req *coretypes.Request, resp *coretypes.Response) error
}

type FileHistory struct{}

func (h *FileHistory) Save(req *coretypes.Request, resp *coretypes.Response) error {
	return nil
}
