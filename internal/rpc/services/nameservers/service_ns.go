// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// NSService 域名服务
type NSService struct {
	services.BaseService
}

// ComposeNSBoard 组合看板数据
func (this *NSService) ComposeNSBoard(ctx context.Context, req *pb.ComposeNSBoardRequest) (*pb.ComposeNSBoardResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var result = &pb.ComposeNSBoardResponse{}

	// 域名
	countDomains, err := nameservers.SharedNSDomainDAO.CountAllEnabledDomains(tx, 0, 0, "")
	if err != nil {
		return nil, err
	}
	result.CountNSDomains = countDomains

	// 记录
	countRecords, err := nameservers.SharedNSRecordDAO.CountAllEnabledRecords(tx)
	if err != nil {
		return nil, err
	}
	result.CountNSRecords = countRecords

	// 集群数
	countClusters, err := nameservers.SharedNSClusterDAO.CountAllEnabledClusters(tx)
	if err != nil {
		return nil, err
	}
	result.CountNSClusters = countClusters

	// 节点数
	countNodes, err := nameservers.SharedNSNodeDAO.CountAllEnabledNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountNSNodes = countNodes

	// 离线节点数
	countOfflineNodes, err := nameservers.SharedNSNodeDAO.CountAllOfflineNodes(tx)
	if err != nil {
		return nil, err
	}
	result.CountOfflineNSNodes = countOfflineNodes

	// 按小时统计
	hourFrom := timeutil.Format("YmdH", time.Now().Add(-23*time.Hour))
	hourTo := timeutil.Format("YmdH")
	hourlyStats, err := nameservers.SharedNSRecordHourlyStatDAO.FindHourlyStats(tx, hourFrom, hourTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range hourlyStats {
		result.HourlyTrafficStats = append(result.HourlyTrafficStats, &pb.ComposeNSBoardResponse_HourlyTrafficStat{
			Hour:          stat.Hour,
			Bytes:         int64(stat.Bytes),
			CountRequests: int64(stat.CountRequests),
		})
	}

	// 按天统计
	dayFrom := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -14))
	dayTo := timeutil.Format("Ymd")
	dailyStats, err := nameservers.SharedNSRecordHourlyStatDAO.FindDailyStats(tx, dayFrom, dayTo)
	if err != nil {
		return nil, err
	}
	for _, stat := range dailyStats {
		result.DailyTrafficStats = append(result.DailyTrafficStats, &pb.ComposeNSBoardResponse_DailyTrafficStat{
			Day:           stat.Day,
			Bytes:         int64(stat.Bytes),
			CountRequests: int64(stat.CountRequests),
		})
	}

	// 域名排行
	topDomainStats, err := nameservers.SharedNSRecordHourlyStatDAO.ListTopDomains(tx, hourFrom, hourTo, 10)
	if err != nil {
		return nil, err
	}
	for _, stat := range topDomainStats {
		domainName, err := nameservers.SharedNSDomainDAO.FindNSDomainName(tx, int64(stat.DomainId))
		if err != nil {
			return nil, err
		}
		if len(domainName) == 0 {
			continue
		}
		result.TopNSDomainStats = append(result.TopNSDomainStats, &pb.ComposeNSBoardResponse_DomainStat{
			NsDomainId:    int64(stat.DomainId),
			NsDomainName:  domainName,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// 节点排行
	topNodeStats, err := nameservers.SharedNSRecordHourlyStatDAO.ListTopNodes(tx, hourFrom, hourTo, 10)
	if err != nil {
		return nil, err
	}
	for _, stat := range topNodeStats {
		nodeName, err := nameservers.SharedNSNodeDAO.FindEnabledNSNodeName(tx, int64(stat.NodeId))
		if err != nil {
			return nil, err
		}
		if len(nodeName) == 0 {
			continue
		}
		result.TopNSNodeStats = append(result.TopNSNodeStats, &pb.ComposeNSBoardResponse_NodeStat{
			NsClusterId:   int64(stat.ClusterId),
			NsNodeId:      int64(stat.NodeId),
			NsNodeName:    nodeName,
			CountRequests: int64(stat.CountRequests),
			Bytes:         int64(stat.Bytes),
		})
	}

	// CPU、内存、负载
	cpuValues, err := models.SharedNodeValueDAO.ListValuesForNSNodes(tx, nodeconfigs.NodeValueItemCPU, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range cpuValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.CpuNodeValues = append(result.CpuNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	memoryValues, err := models.SharedNodeValueDAO.ListValuesForNSNodes(tx, nodeconfigs.NodeValueItemMemory, "usage", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range memoryValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.MemoryNodeValues = append(result.MemoryNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	loadValues, err := models.SharedNodeValueDAO.ListValuesForNSNodes(tx, nodeconfigs.NodeValueItemLoad, "load5m", nodeconfigs.NodeValueRangeMinute)
	if err != nil {
		return nil, err
	}
	for _, v := range loadValues {
		valueJSON, err := json.Marshal(types.Float32(v.Value))
		if err != nil {
			return nil, err
		}
		result.LoadNodeValues = append(result.LoadNodeValues, &pb.NodeValue{
			ValueJSON: valueJSON,
			CreatedAt: int64(v.CreatedAt),
		})
	}

	return result, nil
}
