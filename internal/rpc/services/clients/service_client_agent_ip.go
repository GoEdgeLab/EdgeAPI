// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package clients

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/clients"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ClientAgentIPService Agent IP服务
type ClientAgentIPService struct {
	services.BaseService
}

// CreateClientAgentIPs 创建一组IP
func (this *ClientAgentIPService) CreateClientAgentIPs(ctx context.Context, req *pb.CreateClientAgentIPsRequest) (*pb.RPCSuccess, error) {
	// 先不支持网站服务节点，避免影响普通用户
	_, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.AgentIPs) == 0 {
		return this.Success()
	}

	var tx = this.NullTx()
	for _, agentIP := range req.AgentIPs {
		agentId, err := clients.SharedClientAgentDAO.FindAgentIdWithCode(tx, agentIP.AgentCode)
		if err != nil {
			return nil, err
		}
		if agentId <= 0 {
			continue
		}

		err = clients.SharedClientAgentIPDAO.CreateIP(tx, agentId, agentIP.Ip, agentIP.Ptr)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// ListClientAgentIPsAfterId 查询最新的IP
func (this *ClientAgentIPService) ListClientAgentIPsAfterId(ctx context.Context, req *pb.ListClientAgentIPsAfterIdRequest) (*pb.ListClientAgentIPsAfterIdResponse, error) {
	_, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	if req.Size <= 0 {
		req.Size = 10000
	}

	var tx = this.NullTx()
	var agentMap = map[int64]*clients.ClientAgent{} // agentId => agentCode
	agentIPs, err := clients.SharedClientAgentIPDAO.ListIPsAfterId(tx, req.Id, req.Size)
	if err != nil {
		return nil, err
	}

	var pbIPs = []*pb.ClientAgentIP{}
	for _, agentIP := range agentIPs {
		var agentId = int64(agentIP.AgentId)
		agent, ok := agentMap[agentId]
		if !ok {
			agent, err = clients.SharedClientAgentDAO.FindAgent(tx, agentId)
			if err != nil {
				return nil, err
			}
			if agent == nil {
				continue
			}
			agentMap[agentId] = agent
		}

		pbIPs = append(pbIPs, &pb.ClientAgentIP{
			Id:  int64(agentIP.Id),
			Ip:  agentIP.IP,
			Ptr: "",
			ClientAgent: &pb.ClientAgent{
				Id:          agentId,
				Name:        "",
				Code:        agent.Code,
				Description: "",
			},
		})
	}

	return &pb.ListClientAgentIPsAfterIdResponse{
		ClientAgentIPs: pbIPs,
	}, nil
}
