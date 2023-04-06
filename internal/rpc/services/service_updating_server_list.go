// Copyright 2023 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/serverconfigs"
)

// UpdatingServerListService 待更新服务列表服务
type UpdatingServerListService struct {
	BaseService
}

// FindUpdatingServerLists 查找要更新的服务配置
func (this *UpdatingServerListService) FindUpdatingServerLists(ctx context.Context, req *pb.FindUpdatingServerListsRequest) (*pb.FindUpdatingServerListsResponse, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusterIds, err := models.SharedNodeDAO.FindEnabledAndOnNodeClusterIds(tx, nodeId)
	if err != nil {
		return nil, err
	}

	lists, err := models.SharedUpdatingServerListDAO.FindLists(tx, clusterIds, req.LastId)
	if err != nil {
		return nil, err
	}
	if len(lists) == 0 {
		return &pb.FindUpdatingServerListsResponse{
			MaxId: req.LastId,
		}, nil
	}

	var serverIdMap = map[int64]bool{}
	var serverIds = []int64{}
	var maxId int64
	for _, list := range lists {
		if int64(list.Id) > maxId {
			maxId = int64(list.Id)
		}

		for _, serverId := range list.DecodeServerIds() {
			if !serverIdMap[serverId] {
				serverIdMap[serverId] = true
				serverIds = append(serverIds, serverId)
			}
		}
	}

	if len(serverIds) == 0 {
		return &pb.FindUpdatingServerListsResponse{
			MaxId: req.LastId,
		}, nil
	}

	servers, err := models.SharedServerDAO.FindEnabledServersWithIds(tx, serverIds)
	if err != nil {
		return nil, err
	}
	var serverConfigs = []*serverconfigs.ServerConfig{}
	var cacheMap = utils.NewCacheMap()
	for _, server := range servers {
		serverConfig, err := models.SharedServerDAO.ComposeServerConfig(tx, server, false, nil, cacheMap, true, false)
		if err != nil {
			return nil, err
		}
		if serverConfig == nil {
			continue
		}
		serverConfigs = append(serverConfigs, serverConfig)
	}

	serversJSON, err := json.Marshal(serverConfigs)
	if err != nil {
		return nil, err
	}

	return &pb.FindUpdatingServerListsResponse{
		ServersJSON: serversJSON,
		MaxId:       maxId,
	}, nil
}
