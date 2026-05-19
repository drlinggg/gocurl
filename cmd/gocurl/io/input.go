package io

import (
	"os"

	coreio "github.com/drlinggg/gocurl/core/io"
	coretypes "github.com/drlinggg/gocurl/core/types"
)

var _ coreio.Input = (*cmdInput)(nil)

type cmdInput struct{}

func NewInput() *cmdInput { return &cmdInput{} }

func (cmd *cmdInput) Read() (*coretypes.Request, error) {
	return ParseArgs(os.Args[1:])
}
