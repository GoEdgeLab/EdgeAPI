package node

import (
	"context"
	"github.com/iwind/TeaGo/logs"
)

type Service struct {
}

func (this *Service) Node(context.Context, *NodeRequest) (*NodeResponse, error) {
	logs.Println("you called me")
	return &NodeResponse{}, nil
}
