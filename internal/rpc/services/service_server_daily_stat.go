package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"time"
)

// ServerDailyStatService 服务统计相关服务
type ServerDailyStatService struct {
	BaseService
}

// UploadServerDailyStats 上传统计
func (this *ServerDailyStatService) UploadServerDailyStats(ctx context.Context, req *pb.UploadServerDailyStatsRequest) (*pb.RPCSuccess, error) {
	nodeId, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerDailyStatDAO.SaveStats(tx, req.Stats)
	if err != nil {
		return nil, err
	}

	// 写入其他统计表
	// TODO 将来改成每小时入库一次
	for _, stat := range req.Stats {
		// 总体流量（按天）
		err = stats.SharedTrafficDailyStatDAO.IncreaseDailyBytes(tx, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes)
		if err != nil {
			return nil, err
		}

		// 总体统计（按小时）
		err = stats.SharedTrafficHourlyStatDAO.IncreaseHourlyBytes(tx, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes)
		if err != nil {
			return nil, err
		}

		// 节点流量
		if nodeId > 0 {
			err = stats.SharedNodeTrafficDailyStatDAO.IncreaseDailyBytes(tx, nodeId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes)
			if err != nil {
				return nil, err
			}

			// 集群流量
			clusterId, err := models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
			if err != nil {
				return nil, err
			}
			if clusterId > 0 {
				err = stats.SharedNodeClusterTrafficDailyStatDAO.IncreaseDailyBytes(tx, clusterId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return this.Success()
}

// FindServerHourlyStats 按小时读取统计数据
func (this *ServerDailyStatService) FindServerHourlyStats(ctx context.Context, req *pb.FindServerHourlyStatsRequest) (*pb.FindServerHourlyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.FindServerHourlyStatsResponse_HourlyStat{}
	if req.Hours > 0 {
		for i := int32(0); i < req.Hours; i++ {
			hourString := timeutil.Format("YmdH", time.Now().Add(-time.Duration(i)*time.Hour))
			stat, err := models.SharedServerDailyStatDAO.SumHourlyStat(tx, req.ServerId, hourString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				result = append(result, &pb.FindServerHourlyStatsResponse_HourlyStat{
					Hour:                hourString,
					Bytes:               stat.Bytes,
					CachedBytes:         stat.CachedBytes,
					CountRequests:       stat.CountRequests,
					CountCachedRequests: stat.CountCachedRequests,
				})
			}
		}
	}
	return &pb.FindServerHourlyStatsResponse{Stats: result}, nil
}
