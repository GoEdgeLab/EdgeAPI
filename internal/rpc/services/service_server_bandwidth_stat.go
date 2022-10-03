// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"errors"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"regexp"
	"strings"
	"sync"
	"time"
)

var serverBandwidthStatsMap = map[string]*pb.ServerBandwidthStat{} // key => bandwidth
var serverBandwidthStatsLocker = &sync.Mutex{}

func init() {
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
						// 更新服务的带宽峰值
						if stat.ServerId > 0 {
							err := models.SharedServerBandwidthStatDAO.UpdateServerBandwidth(tx, stat.UserId, stat.ServerId, stat.Day, stat.TimeAt, stat.Bytes)
							if err != nil {
								remotelogs.Error("ServerBandwidthStatService", "dump bandwidth stats failed: "+err.Error())
							}

							err = models.SharedServerDAO.UpdateServerBandwidth(tx, stat.ServerId, stat.Day+stat.TimeAt, stat.Bytes)
							if err != nil {
								remotelogs.Error("ServerBandwidthStatService", "update server bandwidth failed: "+err.Error())
							}
						}

						// 更新服务的带宽峰值
						if stat.UserId > 0 {
							err = models.SharedUserBandwidthStatDAO.UpdateUserBandwidth(tx, stat.UserId, stat.Day, stat.TimeAt, stat.Bytes)
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
func ServerBandwidthCacheKey(serverId int64, day string, timeAt string) string {
	return types.String(serverId) + "@" + day + "@" + timeAt
}

func ServerBandwidthGetCacheBytes(serverId int64, day string, timeAt string) int64 {
	var key = ServerBandwidthCacheKey(serverId, day, timeAt)
	var bytes int64 = 0

	serverBandwidthStatsLocker.Lock()
	stat, ok := serverBandwidthStatsMap[key]
	if ok {
		bytes = stat.Bytes
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
		var key = ServerBandwidthCacheKey(stat.ServerId, stat.Day, stat.TimeAt)
		serverBandwidthStatsLocker.Lock()
		oldStat, ok := serverBandwidthStatsMap[key]
		if ok {
			oldStat.Bytes += stat.Bytes
		} else {
			serverBandwidthStatsMap[key] = &pb.ServerBandwidthStat{
				Id:       0,
				UserId:   stat.UserId,
				ServerId: stat.ServerId,
				Day:      stat.Day,
				TimeAt:   stat.TimeAt,
				Bytes:    stat.Bytes,
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

	var stats = []*models.ServerBandwidthStat{}
	var tx = this.NullTx()
	if len(req.Day) > 0 {
		stats, err = models.SharedServerBandwidthStatDAO.FindAllServerStatsWithDay(tx, req.ServerId, req.Day)
	} else if len(req.Month) > 0 {
		stats, err = models.SharedServerBandwidthStatDAO.FindAllServerStatsWithMonth(tx, req.ServerId, req.Month)
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
	stats, err := models.SharedServerBandwidthStatDAO.FindHourlyBandwidthStats(tx, req.ServerId, req.Hours)
	if err != nil {
		return nil, err
	}

	return &pb.FindHourlyServerBandwidthStatsResponse{
		Stats: stats,
	}, nil
}

// FindDailyServerBandwidthStats 获取最近N天峰值带宽
func (this *ServerBandwidthStatService) FindDailyServerBandwidthStats(ctx context.Context, req *pb.FindDailyServerBandwidthStatsRequest) (*pb.FindDailyServerBandwidthStatsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	stats, err := models.SharedServerBandwidthStatDAO.FindDailyBandwidthStats(tx, req.ServerId, req.Days)
	if err != nil {
		return nil, err
	}

	return &pb.FindDailyServerBandwidthStatsResponse{
		Stats: stats,
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

	if req.UserId <= 0 && req.ServerId <= 0 {
		return &pb.FindDailyServerBandwidthStatsBetweenDaysResponse{
			Stats: nil,
		}, nil
	}

	req.DayFrom = strings.ReplaceAll(req.DayFrom, "-", "")
	req.DayTo = strings.ReplaceAll(req.DayTo, "-", "")

	var dayReg = regexp.MustCompile(`^\d{8}$`)
	if !dayReg.MatchString(req.DayFrom) {
		return nil, errors.New("invalid dayFrom '" + req.DayFrom + "'")
	}
	if !dayReg.MatchString(req.DayTo) {
		return nil, errors.New("invalid dayTo '" + req.DayTo + "'")
	}

	var pbStats = []*pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat{}
	var pbNthStat *pb.FindDailyServerBandwidthStatsBetweenDaysResponse_Stat
	if req.ServerId > 0 { // 服务统计
		pbStats, err = models.SharedServerBandwidthStatDAO.FindBandwidthStatsBetweenDays(tx, req.ServerId, req.DayFrom, req.DayTo)

		// nth
		stat, err := models.SharedServerBandwidthStatDAO.FindPercentileBetweenDays(tx, req.ServerId, req.DayFrom, req.DayTo, req.Percentile)
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
		pbStats, err = models.SharedUserBandwidthStatDAO.FindUserBandwidthStatsBetweenDays(tx, req.UserId, req.DayFrom, req.DayTo)

		// nth
		stat, err := models.SharedUserBandwidthStatDAO.FindPercentileBetweenDays(tx, req.UserId, req.DayFrom, req.DayTo, req.Percentile)
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
