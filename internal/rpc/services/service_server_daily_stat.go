package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"math"
	"regexp"
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

	var tx = this.NullTx()

	// 保存统计数据
	err = models.SharedServerDailyStatDAO.SaveStats(tx, req.Stats)
	if err != nil {
		return nil, err
	}

	var clusterId int64
	switch role {
	case rpcutils.UserTypeDNS:
		clusterId, err = models.SharedNSNodeDAO.FindNodeClusterId(tx, nodeId)
		if err != nil {
			return nil, err
		}
	}

	// 写入其他统计表
	// TODO 将来改成每小时入库一次
	for _, stat := range req.Stats {
		if role == rpcutils.UserTypeNode {
			clusterId, err = models.SharedServerDAO.FindServerClusterId(tx, stat.ServerId)
			if err != nil {
				return nil, err
			}
		}

		// 总体流量（按天）
		err = stats.SharedTrafficDailyStatDAO.IncreaseDailyStat(tx, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}

		// 总体统计（按小时）
		err = stats.SharedTrafficHourlyStatDAO.IncreaseHourlyStat(tx, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}

		// 节点流量
		if nodeId > 0 {
			err = stats.SharedNodeTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
			if err != nil {
				return nil, err
			}

			err = stats.SharedNodeTrafficHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, role, nodeId, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
			if err != nil {
				return nil, err
			}

			// 集群流量
			if clusterId > 0 {
				err = stats.SharedNodeClusterTrafficDailyStatDAO.IncreaseDailyStat(tx, clusterId, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 域名统计
	for _, stat := range req.DomainStats {
		if role == rpcutils.UserTypeNode {
			clusterId, err = models.SharedServerDAO.FindServerClusterId(tx, stat.ServerId)
			if err != nil {
				return nil, err
			}
		}

		err := stats.SharedServerDomainHourlyStatDAO.IncreaseHourlyStat(tx, clusterId, nodeId, stat.ServerId, stat.Domain, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes, stat.CachedBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.AttackBytes)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// FindLatestServerHourlyStats 按小时读取统计数据
func (this *ServerDailyStatService) FindLatestServerHourlyStats(ctx context.Context, req *pb.FindLatestServerHourlyStatsRequest) (*pb.FindLatestServerHourlyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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

// FindServer5MinutelyStatsWithDay 读取某天的5分钟间隔流量
func (this *ServerDailyStatService) FindServer5MinutelyStatsWithDay(ctx context.Context, req *pb.FindServer5MinutelyStatsWithDayRequest) (*pb.FindServer5MinutelyStatsWithDayResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if len(req.Day) == 0 {
		req.Day = timeutil.Format("Ymd")
	}

	dailyStats, err := models.SharedServerDailyStatDAO.FindStatsWithDay(tx, req.ServerId, req.Day, req.TimeFrom, req.TimeTo)
	if err != nil {
		return nil, err
	}

	var pbStats = []*pb.FindServer5MinutelyStatsWithDayResponse_Stat{}
	for _, stat := range dailyStats {
		pbStats = append(pbStats, &pb.FindServer5MinutelyStatsWithDayResponse_Stat{
			Day:                 stat.Day,
			TimeFrom:            stat.TimeFrom,
			TimeTo:              stat.TimeTo,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
		})
	}
	return &pb.FindServer5MinutelyStatsWithDayResponse{Stats: pbStats}, nil
}

// FindLatestServerDailyStats 按天读取统计数据
func (this *ServerDailyStatService) FindLatestServerDailyStats(ctx context.Context, req *pb.FindLatestServerDailyStatsRequest) (*pb.FindLatestServerDailyStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

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

// SumCurrentServerDailyStats 查找单个服务当前统计数据
func (this *ServerDailyStatService) SumCurrentServerDailyStats(ctx context.Context, req *pb.SumCurrentServerDailyStatsRequest) (*pb.SumCurrentServerDailyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx *dbs.Tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 按日
	stat, err := models.SharedServerDailyStatDAO.SumCurrentDailyStat(tx, req.ServerId)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               int64(stat.Bytes),
			CachedBytes:         int64(stat.CachedBytes),
			CountRequests:       int64(stat.CountRequests),
			CountCachedRequests: int64(stat.CountCachedRequests),
			CountAttackRequests: int64(stat.CountAttackRequests),
			AttackBytes:         int64(stat.AttackBytes),
		}
	}

	return &pb.SumCurrentServerDailyStatsResponse{ServerDailyStat: pbStat}, nil
}

// SumServerDailyStats 计算单个服务的日统计
func (this *ServerDailyStatService) SumServerDailyStats(ctx context.Context, req *pb.SumServerDailyStatsRequest) (*pb.SumServerDailyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 某日统计
	var day = timeutil.Format("Ymd")
	if regexp.MustCompile(`^\d{8}$`).MatchString(req.Day) {
		day = req.Day
	}

	stat, err := models.SharedServerDailyStatDAO.SumDailyStat(tx, req.ServerId, day)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               stat.Bytes,
			CachedBytes:         stat.CachedBytes,
			CountRequests:       stat.CountRequests,
			CountCachedRequests: stat.CountCachedRequests,
			CountAttackRequests: stat.CountAttackRequests,
			AttackBytes:         stat.AttackBytes,
		}
	}
	return &pb.SumServerDailyStatsResponse{ServerDailyStat: pbStat}, nil
}

// SumServerMonthlyStats 计算单个服务的月统计
func (this *ServerDailyStatService) SumServerMonthlyStats(ctx context.Context, req *pb.SumServerMonthlyStatsRequest) (*pb.SumServerMonthlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查用户
	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	// 某月统计
	var month = timeutil.Format("Ym")
	if regexp.MustCompile(`^\d{6}$`).MatchString(req.Month) {
		month = req.Month
	}

	// 按月
	stat, err := models.SharedServerDailyStatDAO.SumMonthlyStat(tx, req.ServerId, month)
	if err != nil {
		return nil, err
	}

	var pbStat = &pb.ServerDailyStat{
		ServerId: req.ServerId,
	}
	if stat != nil {
		pbStat = &pb.ServerDailyStat{
			ServerId:            req.ServerId,
			Bytes:               stat.Bytes,
			CachedBytes:         stat.CachedBytes,
			CountRequests:       stat.CountRequests,
			CountCachedRequests: stat.CountCachedRequests,
			CountAttackRequests: stat.CountAttackRequests,
			AttackBytes:         stat.AttackBytes,
		}
	}

	return &pb.SumServerMonthlyStatsResponse{ServerMonthlyStat: pbStat}, nil
}
