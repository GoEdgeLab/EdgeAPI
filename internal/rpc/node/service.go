package node

import (
	"context"
	"github.com/iwind/TeaGo/logs"
)

type Service struct {
}

func (this *Service) Config(ctx context.Context, req *ConfigRequest) (*ConfigResponse, error) {
	logs.Println("you called me")
	return &ConfigResponse{
		Id: req.NodeId,
	}, nil
}
