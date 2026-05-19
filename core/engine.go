package core

import (
	coreio "github.com/drlinggg/gocurl/core/io"
	corestorage "github.com/drlinggg/gocurl/core/storage"
	coretypes "github.com/drlinggg/gocurl/core/types"
)

type Sender interface {
	Send(req *coretypes.HTTPRequest) (*coretypes.Response, error)
}

type Engine struct {
	sender  Sender
	config  corestorage.Config
	history corestorage.History
	input   coreio.Input
	output  coreio.Output
}

func New(sender Sender, config corestorage.Config, history corestorage.History, input coreio.Input, output coreio.Output) *Engine {
	return &Engine{sender: sender, config: config, history: history, input: input, output: output}
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

func (e *Engine) Run() error {
	req, err := e.input.Read()
	if err != nil {
		return err
	}

	resp, err := e.Execute(req)
	if err != nil {
		return err
	}

	return e.output.Write(req, resp)
}
