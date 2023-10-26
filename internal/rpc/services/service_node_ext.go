// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

func (this *NodeService) FindNodeUAMPolicies(ctx context.Context, req *pb.FindNodeUAMPoliciesRequest) (*pb.FindNodeUAMPoliciesResponse, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) FindNodeHTTPCCPolicies(ctx context.Context, req *pb.FindNodeHTTPCCPoliciesRequest) (*pb.FindNodeHTTPCCPoliciesResponse, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) FindNodeHTTP3Policies(ctx context.Context, req *pb.FindNodeHTTP3PoliciesRequest) (*pb.FindNodeHTTP3PoliciesResponse, error) {
	return nil, this.NotImplementedYet()
}

func (this *NodeService) FindNodeHTTPPagesPolicies(ctx context.Context, req *pb.FindNodeHTTPPagesPoliciesRequest) (*pb.FindNodeHTTPPagesPoliciesResponse, error) {
	return nil, this.NotImplementedYet()
}

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

// FindNodeTOAConfig 查找节点的TOA配置
func (this *NodeService) FindNodeTOAConfig(ctx context.Context, req *pb.FindNodeTOAConfigRequest) (*pb.FindNodeTOAConfigResponse, error) {
	return nil, this.NotImplementedYet()
}

// FindNodeNetworkSecurityPolicy 查找节点的网络安全策略
func (this *NodeService) FindNodeNetworkSecurityPolicy(ctx context.Context, req *pb.FindNodeNetworkSecurityPolicyRequest) (*pb.FindNodeNetworkSecurityPolicyResponse, error) {
	return nil, this.NotImplementedYet()
}
