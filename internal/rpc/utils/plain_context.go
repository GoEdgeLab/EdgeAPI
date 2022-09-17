package rpcutils

import (
	"context"
	"time"
)

func IsRest(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	_, ok := ctx.(*PlainContext)
	return ok
}

type PlainContext struct {
	UserType string
	UserId   int64

	ctx context.Context
}

func NewPlainContext(userType string, userId int64) *PlainContext {
	return &PlainContext{
		UserType: userType,
		UserId:   userId,
		ctx:      context.Background(),
	}
}

func (this *PlainContext) Deadline() (deadline time.Time, ok bool) {
	return this.ctx.Deadline()
}

func (this *PlainContext) Done() <-chan struct{} {
	return this.ctx.Done()
}

func (this *PlainContext) Err() error {
	return this.ctx.Err()
}

func (this *PlainContext) Value(key interface{}) interface{} {
	return this.ctx.Value(key)
}
