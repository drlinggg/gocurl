package core

import (
	corestorage "github.com/banakh/gocurl/core/storage"
	coretypes "github.com/banakh/gocurl/core/types"
)

type Sender interface {
	Send(req *coretypes.HTTPRequest) (*coretypes.Response, error)
}

type Engine struct {
	sender  Sender
	config  corestorage.Config
	history corestorage.History
}

func New(sender Sender, config corestorage.Config, history corestorage.History) *Engine {
	return &Engine{sender: sender, config: config, history: history}
}

func (e *Engine) Colors() coretypes.Colors {
	return e.config.Colors()
}

func (e *Engine) Execute(req *coretypes.Request) (*coretypes.Response, error) {
	request, err := e.config.FillRequest(req)
	if err != nil {
		return nil, err
	}

	if err := request.Validate(); err != nil {
		return nil, err
	}

	response, err := e.sender.Send(&request)
	if err != nil {
		return nil, err
	}

	if err := e.history.Save(req, response); err != nil {
		return nil, err
	}

	return response, nil
}
