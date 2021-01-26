package rpcutils

import "context"

type MockNodeContext struct {
	context.Context

	NodeId int64
}

func NewMockNodeContext(nodeId int64) context.Context {
	return &MockNodeContext{NodeId: nodeId}
}
