package rpcutils

import "context"

type MockAdminNodeContext struct {
	context.Context

	AdminId int64
}

func NewMockAdminNodeContext(adminId int64) context.Context {
	return &MockAdminNodeContext{AdminId: adminId}
}
