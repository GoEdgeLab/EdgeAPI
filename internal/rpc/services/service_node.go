package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/pb"
	"github.com/iwind/TeaGo/logs"
)

type NodeService struct {
}

func (this *NodeService) Config(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	logs.Println("you called me")
	return &pb.ConfigResponse{
		Id: req.NodeId,
	}, nil
}
