// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/goman"
	"github.com/TeaOSLab/EdgeAPI/internal/remotelogs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/types"
	"sync"
	"time"
)

var serverBandwidthStatsMap = map[string]*pb.ServerBandwidthStat{} // key => bandwidth
var serverBandwidthStatsLocker = &sync.Mutex{}

func init() {
	var ticker = time.NewTicker(5 * time.Minute)
	if Tea.IsTesting() {
		ticker = time.NewTicker(1 * time.Minute)
	}

	dbs.OnReadyDone(func() {
		goman.New(func() {
			for range ticker.C {
				func() {
					var tx *dbs.Tx

					serverBandwidthStatsLocker.Lock()
					var m = serverBandwidthStatsMap
					serverBandwidthStatsMap = map[string]*pb.ServerBandwidthStat{}
					serverBandwidthStatsLocker.Unlock()

					for _, stat := range m {
						err := models.SharedServerBandwidthStatDAO.UpdateServerBandwidth(tx, stat.UserId, stat.ServerId, stat.Day, stat.TimeAt, stat.Bytes)
						if err != nil {
							remotelogs.Error("ServerBandwidthStatService", "dump bandwidth stats failed: "+err.Error())
						}

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
		var key = types.String(stat.ServerId) + "@" + stat.Day + "@" + stat.TimeAt
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
