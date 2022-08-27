// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"strings"
	"sync"
	"time"
)

// 队列相关数据
var metricStatsMap = map[string]*pb.UploadMetricStatsRequest{} // key (clusterId@nodeId@serverId@itemId) => UploadMetricStatsRequest
var metricStatKeysQueue = make(chan string, 100_000)
var metricStatsLocker = &sync.Mutex{}

func init() {
	dbs.OnReadyDone(func() {
		goman.New(func() {
			// 将队列导入数据库
			var countKeys = 0
			for key := range metricStatKeysQueue {
				err := func(key string) error {
					metricStatsLocker.Lock()
					req, ok := metricStatsMap[key]
					if !ok {
						metricStatsLocker.Unlock()
						return nil
					}
					delete(metricStatsMap, key)
					metricStatsLocker.Unlock()

					var pieces = strings.Split(key, "@")
					var clusterId = types.Int64(pieces[0])
					var nodeId = types.Int64(pieces[1])
					var serverId = types.Int64(pieces[2])
					var itemId = types.Int64(pieces[3])

					// 删除旧的数据
					tx, err := models.SharedMetricStatDAO.Instance.Begin()
					if err != nil {
						return err
					}

					defer func() {
						// 失败时不需要rollback
						commitErr := tx.Commit()
						if commitErr != nil {
							remotelogs.Error("METRIC_STAT", "commit metric stats failed: "+commitErr.Error())
						}
					}()

					err = models.SharedMetricStatDAO.DeleteNodeItemStats(tx, nodeId, serverId, itemId, req.Time)
					if err != nil {
						return err
					}

					for _, stat := range req.MetricStats {
						err := models.SharedMetricStatDAO.CreateStat(tx, stat.Hash, clusterId, nodeId, req.ServerId, req.ItemId, stat.Keys, float64(stat.Value), req.Time, req.Version)
						if err != nil {
							return err
						}
					}

					// 保存总和
					err = models.SharedMetricSumStatDAO.UpdateSum(tx, clusterId, nodeId, req.ServerId, req.Time, req.ItemId, req.Version, req.Count, req.Total)
					if err != nil {
						return err
					}

					return nil
				}(key)
				if err != nil {
					remotelogs.Error("METRIC_STAT", "upload metric stats failed: "+err.Error())
				}

				// 人为限速
				countKeys++
				if countKeys >= 100 {
					countKeys = 0
					time.Sleep(1 * time.Second)
				}
			}
		})
	})
}

// MetricStatService 指标统计数据相关服务
type MetricStatService struct {
	BaseService
}

// UploadMetricStats 上传统计数据
func (this *MetricStatService) UploadMetricStats(ctx context.Context, req *pb.UploadMetricStatsRequest) (*pb.RPCSuccess, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	clusterId, err := models.SharedServerDAO.FindServerClusterId(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var key = types.String(clusterId) + "@" + types.String(nodeId) + "@" + types.String(req.ServerId) + "@" + types.String(req.ItemId)
	metricStatsLocker.Lock()
	metricStatsMap[key] = req

	select {
	case metricStatKeysQueue <- key:
	default:
		// 如果满了就删除
		delete(metricStatsMap, key)
	}

	metricStatsLocker.Unlock()

	return this.Success()
}

// CountMetricStats 计算指标数据数量
func (this *MetricStatService) CountMetricStats(ctx context.Context, req *pb.CountMetricStatsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	version, err := models.SharedMetricItemDAO.FindItemVersion(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	count, err := models.SharedMetricStatDAO.CountItemStats(tx, req.MetricItemId, version)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListMetricStats 读取单页指标数据
func (this *MetricStatService) ListMetricStats(ctx context.Context, req *pb.ListMetricStatsRequest) (*pb.ListMetricStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	version, err := models.SharedMetricItemDAO.FindItemVersion(tx, req.MetricItemId)
	if err != nil {
		return nil, err
	}
	stats, err := models.SharedMetricStatDAO.ListItemStats(tx, req.MetricItemId, version, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbStats []*pb.MetricStat
	for _, stat := range stats {
		// cluster
		clusterName, err := models.SharedNodeClusterDAO.FindNodeClusterName(tx, int64(stat.ClusterId))
		if err != nil {
			return nil, err
		}

		// node
		nodeName, err := models.SharedNodeDAO.FindNodeName(tx, int64(stat.NodeId))
		if err != nil {
			return nil, err
		}

		// server
		serverName, err := models.SharedServerDAO.FindEnabledServerName(tx, int64(stat.ServerId))
		if err != nil {
			return nil, err
		}

		// 查找sum值
		count, total, err := models.SharedMetricSumStatDAO.FindNodeServerSum(tx, int64(stat.NodeId), int64(stat.ServerId), stat.Time, int64(stat.ItemId), types.Int32(stat.Version))
		if err != nil {
			return nil, err
		}

		pbStats = append(pbStats, &pb.MetricStat{
			Id:          int64(stat.Id),
			Hash:        stat.Hash,
			ServerId:    int64(stat.ServerId),
			ItemId:      int64(stat.ItemId),
			Keys:        stat.DecodeKeys(),
			Value:       float32(stat.Value),
			Time:        stat.Time,
			Version:     types.Int32(stat.Version),
			NodeCluster: &pb.NodeCluster{Id: int64(stat.ClusterId), Name: clusterName},
			Node:        &pb.Node{Id: int64(stat.NodeId), Name: nodeName},
			Server:      &pb.Server{Id: int64(stat.ServerId), Name: serverName},
			SumCount:    count,
			SumTotal:    total,
		})
	}
	return &pb.ListMetricStatsResponse{MetricStats: pbStats}, nil
}
