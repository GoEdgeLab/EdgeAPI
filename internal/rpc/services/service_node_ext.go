// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

func (this *NodeService) FindNodeScheduleInfo(ctx context.Context, req *pb.FindNodeScheduleInfoRequest) (*pb.FindNodeScheduleInfoResponse, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) UpdateNodeScheduleInfo(ctx context.Context, req *pb.UpdateNodeScheduleInfoRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) ResetNodeActionStatus(ctx context.Context, req *pb.ResetNodeActionStatusRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) FindAllNodeScheduleInfoWithNodeClusterId(ctx context.Context, req *pb.FindAllNodeScheduleInfoWithNodeClusterIdRequest) (*pb.FindAllNodeScheduleInfoWithNodeClusterIdResponse, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) CopyNodeActionsToNodeGroup(ctx context.Context, req *pb.CopyNodeActionsToNodeGroupRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) CopyNodeActionsToNodeCluster(ctx context.Context, req *pb.CopyNodeActionsToNodeClusterRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}
