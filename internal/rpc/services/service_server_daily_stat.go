package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"math"
	"time"
)

// ServerDailyStatService 服务统计相关服务
type ServerDailyStatService struct {
	BaseService
}

// UploadServerDailyStats 上传统计
func (this *ServerDailyStatService) UploadServerDailyStats(ctx context.Context, req *pb.UploadServerDailyStatsRequest) (*pb.RPCSuccess, error) {
	role, nodeId, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedServerDailyStatDAO.SaveStats(tx, req.Stats)
	if err != nil {
		return nil, err
	}

	var clusterId int64
	switch role {
	case rpcutils.UserTypeNode:
		clusterId, err = models.SharedNodeDAO.FindNodeClusterId(tx, nodeId)
		if err != nil {
			return nil, err
		}
	case rpcutils.UserTypeDNS:
		clusterId, err = nameservers.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
		if err != nil {
			return nil, err
		}
	}

	// 写入其他统计表
	// TODO 将来改成每小时入库一次
	for _, stat := range req.Stats {
		// 总体流量（按天）
		err = stats.SharedTrafficDailyStatDAO.IncreaseDailyStat(tx, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
		if err != nil {
			return nil, err
		}

		// 总体统计（按小时）
		err = stats.SharedTrafficHourlyStatDAO.IncreaseHourlyStat(tx, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
		if err != nil {
			return nil, err
		}

		// 节点流量
		if nodeId > 0 {
			err = stats.SharedNodeTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
			if err != nil {
				return nil, err
			}

			err = stats.SharedNodeTrafficHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
			if err != nil {
				return nil, err
			}

			// 集群流量
			if clusterId > 0 {
				err = stats.SharedNodeClusterTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 域名统计
	for _, stat := range req.DomainStats {
		err := stats.SharedServerDomainHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, nodeId, stat.ServerId, stat.Domain, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindLatestServerHourlyStats 按小时读取统计数据
func (this *ServerDailyStatService) FindLatestServerHourlyStats(ctx context.Context, req *pb.FindLatestServerHourlyStatsRequest) (*pb.FindLatestServerHourlyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.FindLatestServerHourlyStatsResponse_HourlyStat{}
	if req.Hours > 0 {
		for i := int32(0); i < req.Hours; i++ {
			hourString := timeutil.Format("YmdH", time.Now().Add(-time.Duration(i)*time.Hour))
			stat, err := models.SharedServerDailyStatDAO.SumHourlyStat(tx, req.ServerId, hourString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				result = append(result, &pb.FindLatestServerHourlyStatsResponse_HourlyStat{
					Hour:                hourString,
					Bytes:               stat.Bytes,
					CachedBytes:         stat.CachedBytes,
					CountRequests:       stat.CountRequests,
					CountCachedRequests: stat.CountCachedRequests,
				})
			}
		}
	}
	return &pb.FindLatestServerHourlyStatsResponse{Stats: result}, nil
}

// FindLatestServerMinutelyStats 按分钟读取统计数据
func (this *ServerDailyStatService) FindLatestServerMinutelyStats(ctx context.Context, req *pb.FindLatestServerMinutelyStatsRequest) (*pb.FindLatestServerMinutelyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{}
	cache := map[string]*pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{} // minute => stat

	var avgRatio int64 = 5 * 60 // 因为每5分钟记录一次

	if req.Minutes > 0 {
		for i := int32(0); i < req.Minutes; i++ {
			date := time.Now().Add(-time.Duration(i) * time.Minute)
			minuteString := timeutil.Format("YmdHi", date)

			minute := date.Minute()
			roundMinute := minute - minute%5
			if roundMinute != minute {
				date = date.Add(-time.Duration(minute-roundMinute) * time.Minute)
			}
			queryMinuteString := timeutil.Format("YmdHi", date)
			pbStat, ok := cache[queryMinuteString]
			if ok {
				result = append(result, &pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{
					Minute:              minuteString,
					Bytes:               pbStat.Bytes,
					CachedBytes:         pbStat.CachedBytes,
					CountRequests:       pbStat.CountRequests,
					CountCachedRequests: pbStat.CountCachedRequests,
				})
				continue
			}

			stat, err := models.SharedServerDailyStatDAO.SumMinutelyStat(tx, req.ServerId, queryMinuteString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				pbStat = &pb.FindLatestServerMinutelyStatsResponse_MinutelyStat{
					Minute:              minuteString,
					Bytes:               stat.Bytes / avgRatio,
					CachedBytes:         stat.CachedBytes / avgRatio,
					CountRequests:       int64(math.Ceil(float64(stat.CountRequests) / float64(avgRatio))),
					CountCachedRequests: int64(math.Ceil(float64(stat.CountCachedRequests) / float64(avgRatio))),
				}
				result = append(result, pbStat)
				cache[queryMinuteString] = pbStat
			}
		}
	}
	return &pb.FindLatestServerMinutelyStatsResponse{Stats: result}, nil
}

// FindLatestServerDailyStats 按天读取统计数据
func (this *ServerDailyStatService) FindLatestServerDailyStats(ctx context.Context, req *pb.FindLatestServerDailyStatsRequest) (*pb.FindLatestServerDailyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	result := []*pb.FindLatestServerDailyStatsResponse_DailyStat{}
	if req.Days > 0 {
		for i := int32(0); i < req.Days; i++ {
			dayString := timeutil.Format("Ymd", time.Now().AddDate(0, 0, -int(i)))
			stat, err := models.SharedServerDailyStatDAO.SumDailyStat(tx, req.ServerId, dayString)
			if err != nil {
				return nil, err
			}
			if stat != nil {
				result = append(result, &pb.FindLatestServerDailyStatsResponse_DailyStat{
					Day:                 dayString,
					Bytes:               stat.Bytes,
					CachedBytes:         stat.CachedBytes,
					CountRequests:       stat.CountRequests,
					CountCachedRequests: stat.CountCachedRequests,
				})
			}
		}
	}
	return &pb.FindLatestServerDailyStatsResponse{Stats: result}, nil
}
