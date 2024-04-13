// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"encoding/json"
	"errors"
	teaconst "github.com/TeaOSLab/EdgeAPI/internal/const"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/events"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeAPI/internal/utils/regexputils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/TeaOSLab/EdgeCommon/pkg/systemconfigs"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	timeutil "github.com/iwind/TeaGo/utils/time"
	"os"
	"strings"
	"sync"
	"time"
)

var serverBandwidthStatsMap = map[string]*pb.ServerBandwidthStat{} // server key => bandwidth
var serverBandwidthStatsLocker = &sync.Mutex{}

func init() {
	// 数据缓存
	if teaconst.IsMain {
		var cacheFile = Tea.Root + "/data/server_bandwidth_stats.cache"

		{
			data, err := os.ReadFile(cacheFile)
			if err == nil {
				_ = os.Remove(cacheFile)
				serverBandwidthStatsLocker.Lock()
				_ = json.Unmarshal(data, &serverBandwidthStatsMap)
				serverBandwidthStatsLocker.Unlock()
			}
		}

		events.On(events.EventQuit, func() {
			serverBandwidthStatsMapJSON, err := json.Marshal(serverBandwidthStatsMap)
			if err == nil {
				_ = os.WriteFile(cacheFile, serverBandwidthStatsMapJSON, 0666)
			}
		})
	}

	// 定时处理数据
	var ticker = time.NewTicker(1 * time.Minute)
	var useTx = true

	dbs.OnReadyDone(func() {
		goman.New(func() {
			for range ticker.C {
				func() {
					serverBandwidthStatsLocker.Lock()
					var m = serverBandwidthStatsMap
					serverBandwidthStatsMap = map[string]*pb.ServerBandwidthStat{}
					serverBandwidthStatsLocker.Unlock()

					var tx *dbs.Tx
					var err error

					if useTx {
						var before = time.Now()

						tx, err = models.SharedServerBandwidthStatDAO.Instance.Begin()
						if err != nil {
							remotelogs.Error("ServerBandwidthStatService", "begin transaction failed: "+err.Error())
							return
						}

						defer func() {
							if tx != nil {
								commitErr := tx.Commit()
								if commitErr != nil {
									remotelogs.Error("METRIC_STAT", "commit bandwidth stats failed: "+commitErr.Error())
								}
							}

							// 如果运行时间过长，则不使用事务
							if time.Since(before) > 1*time.Second {
								useTx = false
							}
						}()
					}

					for _, stat := range m {
						// 更新网站的带宽峰值
						if stat.ServerId > 0 {
							// 更新带宽统计
							err = models.SharedServerBandwidthStatDAO.UpdateServerBandwidth(tx, stat.UserId, stat.ServerId, stat.NodeRegionId, stat.UserPlanId, stat.Day, stat.TimeAt, stat.Bytes, stat.TotalBytes, stat.CachedBytes, stat.AttackBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.CountIPs)
							if err != nil {
								remotelogs.Error("ServerBandwidthStatService", "dump bandwidth stats failed: "+err.Error())
							}

							// 更新网站的bandwidth字段，方便快速排序
							err = models.SharedServerDAO.UpdateServerBandwidth(tx, stat.ServerId, stat.Day+stat.TimeAt, stat.Bytes, stat.CountRequests, stat.CountAttackRequests)
							if err != nil {
								remotelogs.Error("ServerBandwidthStatService", "update server bandwidth failed: "+err.Error())
							}

							// 套餐统计
							if stat.UserPlanId > 0 {
								// 总体统计
								err = models.SharedUserPlanStatDAO.IncreaseUserPlanStat(tx, stat.UserPlanId, stat.TotalBytes, stat.CountRequests, stat.CountWebsocketConnections)
								if err != nil {
									remotelogs.Error("ServerBandwidthStatService", "IncreaseUserPlanStat: "+err.Error())
								}

								// 分时统计
								err = models.SharedUserPlanBandwidthStatDAO.UpdateUserPlanBandwidth(tx, stat.UserId, stat.UserPlanId, stat.NodeRegionId, stat.Day, stat.TimeAt, stat.Bytes, stat.TotalBytes, stat.CachedBytes, stat.AttackBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests, stat.CountWebsocketConnections)
							}
						}

						// 更新用户的带宽峰值
						if stat.UserId > 0 {
							err = models.SharedUserBandwidthStatDAO.UpdateUserBandwidth(tx, stat.UserId, stat.NodeRegionId, stat.Day, stat.TimeAt, stat.Bytes, stat.TotalBytes, stat.CachedBytes, stat.AttackBytes, stat.CountRequests, stat.CountCachedRequests, stat.CountAttackRequests)
							if err != nil {
								remotelogs.Error("SharedUserBandwidthStatDAO", "dump bandwidth stats failed: "+err.Error())
							}
						}
					}
				}()
			}
		})
	})
}

// ServerBandwidthCacheKey 组合缓存Key
func ServerBandwidthCacheKey(serverId int64, regionId int64, day string, timeAt string) string {
	return types.String(serverId) + "@" + types.String(regionId) + "@" + day + "@" + timeAt
}

func ServerBandwidthGetCacheBytes(serverId int64, timeAt string) int64 {
	var bytes int64 = 0

	serverBandwidthStatsLocker.Lock()
	for _, stat := range serverBandwidthStatsMap {
		if stat.ServerId == serverId && stat.TimeAt == timeAt {
			bytes += stat.Bytes
		}
	}
	serverBandwidthStatsLocker.Unlock()

	return bytes
}

type ServerBandwidthStatService struct {
	BaseService
}

// UploadServerBandwidthStats 上传带宽统计
func (this *ServerBandwidthStatService) UploadServerBandwidthStats(ctx context.Context, req *pb.UploadServerBandwidthStatsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateNode(ctx)
	if err != nil {
		return nil, err
	}

	for _, stat := range req.ServerBandwidthStats {
		var key = ServerBandwidthCacheKey(stat.ServerId, stat.NodeRegionId, stat.Day, stat.TimeAt)
		serverBandwidthStatsLocker.Lock()
		oldStat, ok := serverBandwidthStatsMap[key]
		if ok {
			oldStat.Bytes += stat.Bytes
			oldStat.TotalBytes += stat.TotalBytes
			oldStat.CachedBytes += stat.CachedBytes
			oldStat.AttackBytes += stat.AttackBytes
			oldStat.CountRequests += stat.CountRequests
			oldStat.CountCachedRequests += stat.CountCachedRequests
			oldStat.CountAttackRequests += stat.CountAttackRequests
			oldStat.CountWebsocketConnections += stat.CountWebsocketConnections
			oldStat.CountIPs += stat.CountIPs
		} else {
			serverBandwidthStatsMap[key] = &pb.ServerBandwidthStat{
				Id:                        0,
				NodeRegionId:              stat.NodeRegionId,
				UserId:                    stat.UserId,
				ServerId:                  stat.ServerId,
				Day:                       stat.Day,
				TimeAt:                    stat.TimeAt,
				Bytes:                     stat.Bytes,
				TotalBytes:                stat.TotalBytes,
				CachedBytes:               stat.CachedBytes,
				AttackBytes:               stat.AttackBytes,
				CountRequests:             stat.CountRequests,
				CountCachedRequests:       stat.CountCachedRequests,
				CountAttackRequests:       stat.CountAttackRequests,
				CountWebsocketConnections: stat.CountWebsocketConnections,
				UserPlanId:                stat.UserPlanId,
				CountIPs:                  stat.CountIPs,
			}
		}
		serverBandwidthStatsLocker.Unlock()
	}

	return this.Success()
}

// FindServerBandwidthStats 获取服务的峰值带宽
func (this *ServerBandwidthStatService) FindServerBandwidthStats(ctx context.Context, req *pb.FindServerBandwidthStatsRequest) (*pb.FindServerBandwidthStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 带宽算法
	if len(req.Algo) == 0 {
		userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		bandwidthAlgo, err := models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, userId, nil)
		if err != nil {
			return nil, err
		}
		req.Algo = bandwidthAlgo
	}

	var stats []*models.ServerBandwidthStat
	if len(req.Day) > 0 {
		stats, err = models.SharedServerBandwidthStatDAO.FindAllServerStatsWithDay(tx, req.ServerId, req.Day, req.Algo == systemconfigs.BandwidthAlgoAvg)
	} else if len(req.Month) > 0 {
		stats, err = models.SharedServerBandwidthStatDAO.FindAllServerStatsWithMonth(tx, req.ServerId, req.Month, req.Algo == systemconfigs.BandwidthAlgoAvg)
	} else {
		// 默认返回空
		return nil, errors.New("'month' or 'day' parameter is needed")
	}

	if err != nil {
		return nil, err
	}

	var pbStats = []*pb.ServerBandwidthStat{}
	for _, stat := range stats {
		pbStats = append(pbStats, &pb.ServerBandwidthStat{
			Id:       int64(stat.Id),
			UserId:   int64(stat.UserId),
			ServerId: int64(stat.ServerId),
			Day:      stat.Day,
			TimeAt:   stat.TimeAt,
			Bytes:    int64(stat.Bytes),
		})
	}
	return &pb.FindServerBandwidthStatsResponse{
		ServerBandwidthStats: pbStats,
	}, nil
}

// FindHourlyServerBandwidthStats 获取最近N小时峰值带宽
func (this *ServerBandwidthStatService) FindHourlyServerBandwidthStats(ctx context.Context, req *pb.FindHourlyServerBandwidthStatsRequest) (*pb.FindHourlyServerBandwidthStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 带宽算法
	if len(req.Algo) == 0 {
		userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		bandwidthAlgo, err := models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, userId, nil)
		if err != nil {
			return nil, err
		}
		req.Algo = bandwidthAlgo
	}

	if req.Hours <= 0 {
		req.Hours = 12
	}

	stats, err := models.SharedServerBandwidthStatDAO.FindHourlyBandwidthStats(tx, req.ServerId, req.Hours, req.Algo == systemconfigs.BandwidthAlgoAvg)
	if err != nil {
		return nil, err
	}

	// percentile
	var percentile = systemconfigs.DefaultBandwidthPercentile
	userUIConfig, _ := models.SharedSysSettingDAO.ReadUserUIConfig(tx)
	if userUIConfig != nil && userUIConfig.TrafficStats.BandwidthPercentile > 0 {
		percentile = userUIConfig.TrafficStats.BandwidthPercentile
	}

	var timestamp = time.Now().Unix() - int64(req.Hours)*3600
	var timeFrom = timeutil.FormatTime("YmdH00", timestamp)
	var timeTo = timeutil.Format("YmdHi")

	var pbNthStat *pb.FindHourlyServerBandwidthStatsResponse_Stat
	percentileStat, err := models.SharedServerBandwidthStatDAO.FindPercentileBetweenTimes(tx, req.ServerId, timeFrom, timeTo, percentile, req.Algo == systemconfigs.BandwidthAlgoAvg)
	if err != nil {
		return nil, err
	}
	if percentileStat != nil {
		pbNthStat = &pb.FindHourlyServerBandwidthStatsResponse_Stat{
			Day:   percentileStat.Day,
			Hour:  types.Int32(percentileStat.TimeAt[:2]),
			Bytes: int64(percentileStat.Bytes),
			Bits:  int64(percentileStat.Bytes * 8),
		}
	}

	return &pb.FindHourlyServerBandwidthStatsResponse{
		Stats:      stats,
		Percentile: percentile,
		NthStat:    pbNthStat,
	}, nil
}

// FindDailyServerBandwidthStats 获取最近N天峰值带宽
func (this *ServerBandwidthStatService) FindDailyServerBandwidthStats(ctx context.Context, req *pb.FindDailyServerBandwidthStatsRequest) (*pb.FindDailyServerBandwidthStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 带宽算法
	if len(req.Algo) == 0 {
		userId, err := models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
		if err != nil {
			return nil, err
		}
		bandwidthAlgo, err := models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, userId, nil)
		if err != nil {
			return nil, err
		}
		req.Algo = bandwidthAlgo
	}

	if req.Days <= 0 {
		req.Days = 30
	}

	var timestamp = time.Now().Unix() - int64(req.Days)*86400
	var dayFrom = timeutil.FormatTime("Ymd", timestamp)
	var dayTo = timeutil.Format("Ymd")

	stats, err := models.SharedServerBandwidthStatDAO.FindBandwidthStatsBetweenDays(tx, req.ServerId, dayFrom, dayTo, req.Algo == systemconfigs.BandwidthAlgoAvg)
	if err != nil {
		return nil, err
	}
	var pbStats = []*pb.FindDailyServerBandwidthStatsResponse_Stat{}
	for _, stat := range stats {
		pbStats = append(pbStats, &pb.FindDailyServerBandwidthStatsResponse_Stat{
			Day:   stat.Day,
			Bytes: stat.Bytes,
			Bits:  stat.Bytes * 8,
		})
	}

	// percentile
	var percentile = systemconfigs.DefaultBandwidthPercentile
	userUIConfig, _ := models.SharedSysSettingDAO.ReadUserUIConfig(tx)
	if userUIConfig != nil && userUIConfig.TrafficStats.BandwidthPercentile > 0 {
		percentile = userUIConfig.TrafficStats.BandwidthPercentile
	}

	var pbNthStat = &pb.FindDailyServerBandwidthStatsResponse_Stat{}
	percentileStat, err := models.SharedServerBandwidthStatDAO.FindPercentileBetweenDays(tx, req.ServerId, dayFrom, dayTo, percentile, req.Algo == systemconfigs.BandwidthAlgoAvg)
	if err != nil {
		return nil, err
	}
	if percentileStat != nil {
		pbNthStat = &pb.FindDailyServerBandwidthStatsResponse_Stat{
			Day:   percentileStat.Day,
			Bytes: int64(percentileStat.Bytes),
			Bits:  int64(percentileStat.Bytes * 8),
		}
	}

	return &pb.FindDailyServerBandwidthStatsResponse{
		Stats:      pbStats,
		Percentile: percentile,
		NthStat:    pbNthStat,
	}, nil
}

// FindDailyServerBandwidthStatsBetweenDays 读取日期段内的带宽数据
func (this *ServerBandwidthStatService) FindDailyServerBandwidthStatsBetweenDays(ctx context.Context, req *pb.FindDailyServerBandwidthStatsBetweenDaysRequest) (*pb.FindDailyServerBandwidthStatsBetweenDaysResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	if userId > 0 {
		req.UserId = userId

		// 检查权限
		if req.ServerId > 0 {
			err = models.SharedServerDAO.CheckUserServer(tx, userId, req.ServerId)
			if err != nil {
				return nil, err
			}
		}
	}

	// 带宽算法
	if len(req.Algo) == 0 {
		var bandwidthUserId = userId
		if bandwidthUserId <= 0 {
			if req.UserId > 0 {
				bandwidthUserId = req.UserId
			} else if req.ServerId > 0 {
				bandwidthUserId, err = models.SharedServerDAO.FindServerUserId(tx, req.ServerId)
				if err != nil {
					return nil, err
				}
			}
		}
		if bandwidthUserId > 0 {
			req.Algo, err = models.SharedUserDAO.FindUserBandwidthAlgoForView(tx, bandwidthUserId, nil)
			if err != nil {
				return nil, err
			}
		}
	}

	if req.UserId <= 0 && req.ServerId <= 0 {
		return &pb.FindDailyServerBandwidthStatsBetweenDaysResponse{
			Stats: nil,
		}, nil
	}

	req.DayFrom = strings.ReplaceAll(req.DayFrom, "-", "")
	req.DayTo = strings.ReplaceAll(req.DayTo, "-", "")

	if !regexputils.YYYYMMDD.MatchString(req.DayFrom) {
		return nil, errors.New("invalid dayFrom '" + req.DayFrom + "'")
	}
	if !regexputils.YYYYMMDD.MatchString(req.DayTo) {
		return nil, errors.New("invalid dayTo '" + req.DayTo + "'")
	}

	var pbStats []*pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat
	var pbNthStat *pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat
	if req.ServerId > 0 { // 服务统计
		pbStats, err = models.SharedServerBandwidthStatDAO.FindBandwidthStatsBetweenDays(tx, req.ServerId, req.DayFrom, req.DayTo, req.Algo == systemconfigs.BandwidthAlgoAvg)

		// nth
		stat, err := models.SharedServerBandwidthStatDAO.FindPercentileBetweenDays(tx, req.ServerId, req.DayFrom, req.DayTo, req.Percentile, req.Algo == systemconfigs.BandwidthAlgoAvg)
		if err != nil {
			return nil, err
		}
		if stat != nil {
			pbNthStat = &pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{
				Day:    stat.Day,
				TimeAt: stat.TimeAt,
				Bytes:  int64(stat.Bytes),
				Bits:   int64(stat.Bytes * 8),
			}
		}
	} else { // 用户统计
		pbStats, err = models.SharedUserBandwidthStatDAO.FindUserBandwidthStatsBetweenDays(tx, req.UserId, req.NodeRegionId, req.DayFrom, req.DayTo, req.Algo == systemconfigs.BandwidthAlgoAvg)

		// nth
		stat, err := models.SharedUserBandwidthStatDAO.FindPercentileBetweenDays(tx, req.UserId, req.NodeRegionId, req.DayFrom, req.DayTo, req.Percentile, req.Algo == systemconfigs.BandwidthAlgoAvg)
		if err != nil {
			return nil, err
		}
		if stat != nil {
			pbNthStat = &pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{
				Day:    stat.Day,
				TimeAt: stat.TimeAt,
				Bytes:  int64(stat.Bytes),
				Bits:   int64(stat.Bytes * 8),
			}
		}
	}
	if err != nil {
		return nil, err
	}

	return &pb.FindDailyServerBandwidthStatsBetweenDaysResponse{
		Stats:   pbStats,
		NthStat: pbNthStat,
	}, nil
}
