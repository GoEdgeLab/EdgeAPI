// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
)

// HTTPCacheTaskKeyService 缓存任务Key管理
type HTTPCacheTaskKeyService struct {
	BaseService
}

// ValidateHTTPCacheTaskKeys 校验缓存Key
func (this *HTTPCacheTaskKeyService) ValidateHTTPCacheTaskKeys(ctx context.Context, req *pb.ValidateHTTPCacheTaskKeysRequest) (*pb.ValidateHTTPCacheTaskKeysResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx

	// 检查Key数量
	var clusterId int64
	if userId > 0 {
		// TODO 限制当日刷新总条数（配额）
		if len(req.Keys) > models.MaxKeysPerTask {
			return nil, errors.New("too many keys (current:" + types.String(len(req.Keys)) + ", max:" + types.String(models.MaxKeysPerTask) + ")")
		}

		clusterId, err = models.SharedUserDAO.FindUserClusterId(tx, userId)
		if err != nil {
			return nil, err
		}
	}

	var pbFailResults = []*pb.ValidateHTTPCacheTaskKeysResponse_FailKey{}
	var domainMap = map[string]*models.Server{} // domain name => *Server
	for _, key := range req.Keys {
		if len(key) == 0 {
			pbFailResults = append(pbFailResults, &pb.ValidateHTTPCacheTaskKeysResponse_FailKey{
				Key:        key,
				ReasonCode: "requireKey",
			})
			continue
		}

		// 获取域名
		var domain = utils.ParseDomainFromKey(key)
		if len(domain) == 0 {
			pbFailResults = append(pbFailResults, &pb.ValidateHTTPCacheTaskKeysResponse_FailKey{
				Key:        key,
				ReasonCode: "requireDomain",
			})
			continue
		}

		// 查询所在集群
		server, ok := domainMap[domain]
		if !ok {
			server, err = models.SharedServerDAO.FindEnabledServerWithDomain(tx, domain)
			if err != nil {
				return nil, err
			}
			if server == nil {
				pbFailResults = append(pbFailResults, &pb.ValidateHTTPCacheTaskKeysResponse_FailKey{
					Key:        key,
					ReasonCode: "requireServer",
				})
				continue
			}
			domainMap[domain] = server
		}

		// 检查用户
		if userId > 0 {
			if int64(server.UserId) != userId {
				pbFailResults = append(pbFailResults, &pb.ValidateHTTPCacheTaskKeysResponse_FailKey{
					Key:        key,
					ReasonCode: "requireUser",
				})
				continue
			}
		}

		var serverClusterId = int64(server.ClusterId)
		if serverClusterId == 0 {
			if clusterId > 0 {
				serverClusterId = clusterId
			} else {
				pbFailResults = append(pbFailResults, &pb.ValidateHTTPCacheTaskKeysResponse_FailKey{
					Key:        key,
					ReasonCode: "requireClusterId",
				})
				continue
			}
		}
	}

	return &pb.ValidateHTTPCacheTaskKeysResponse{FailKeys: pbFailResults}, nil
}

// FindDoingHTTPCacheTaskKeys 查找需要执行的Key
func (this *HTTPCacheTaskKeyService) FindDoingHTTPCacheTaskKeys(ctx context.Context, req *pb.FindDoingHTTPCacheTaskKeysRequest) (*pb.FindDoingHTTPCacheTaskKeysResponse, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	if req.Size <= 0 {
		req.Size = 100
	}

	var tx *dbs.Tx
	keys, err := models.SharedHTTPCacheTaskKeyDAO.FindDoingTaskKeys(tx, nodeId, req.Size)
	if err != nil {
		return nil, err
	}

	var pbKeys = []*pb.HTTPCacheTaskKey{}
	for _, key := range keys {
		pbKeys = append(pbKeys, &pb.HTTPCacheTaskKey{
			Id:            int64(key.Id),
			TaskId:        int64(key.TaskId),
			Key:           key.Key,
			Type:          key.Type,
			KeyType:       key.KeyType,
			NodeClusterId: int64(key.ClusterId),
		})
	}

	return &pb.FindDoingHTTPCacheTaskKeysResponse{HttpCacheTaskKeys: pbKeys}, nil
}

// UpdateHTTPCacheTaskKeysStatus 更新一组Key状态
func (this *HTTPCacheTaskKeyService) UpdateHTTPCacheTaskKeysStatus(ctx context.Context, req *pb.UpdateHTTPCacheTaskKeysStatusRequest) (*pb.RPCSuccess, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx

	var nodesJSONMap = map[int64][]byte{} // clusterId => nodesJSON

	for _, result := range req.KeyResults {
		// 集群Id
		var clusterId = result.NodeClusterId
		nodesJSON, ok := nodesJSONMap[clusterId]
		if !ok {
			nodeIdsInCluster, err := models.SharedNodeDAO.FindEnabledAndOnNodeIdsWithClusterId(tx, clusterId, true)
			if err != nil {
				return nil, err
			}
			var nodeMap = map[int64]bool{}
			for _, nodeIdInCluster := range nodeIdsInCluster {
				nodeMap[nodeIdInCluster] = true
			}
			nodesJSON, err = json.Marshal(nodeMap)
			if err != nil {
				return nil, err
			}
			nodesJSONMap[clusterId] = nodesJSON
		}

		err = models.SharedHTTPCacheTaskKeyDAO.UpdateKeyStatus(tx, result.Id, nodeId, result.Error, nodesJSON)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}
