// Copyright 2023 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .
//go:build !plus

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

func (this *NodeClusterService) FindNodeClusterHTTP3Policy(ctx context.Context, req *pb.FindNodeClusterHTTP3PolicyRequest) (*pb.FindNodeClusterHTTP3PolicyResponse, error) {
	return nil, this.NotImplementedYet()
}

// FindNodeClusterNetworkSecurityPolicy 获取集群的网络安全策略
func (this *NodeClusterService) FindNodeClusterNetworkSecurityPolicy(ctx context.Context, req *pb.FindNodeClusterNetworkSecurityPolicyRequest) (*pb.FindNodeClusterNetworkSecurityPolicyResponse, error) {
	return nil, this.NotImplementedYet()
}

// UpdateNodeClusterNetworkSecurityPolicy 修改集群的网络安全策略
func (this *NodeClusterService) UpdateNodeClusterNetworkSecurityPolicy(ctx context.Context, req *pb.UpdateNodeClusterNetworkSecurityPolicyRequest) (*pb.RPCSuccess, error) {
	return nil, this.NotImplementedYet()
}
