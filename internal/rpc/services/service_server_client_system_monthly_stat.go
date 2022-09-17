package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/stats"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ServerClientSystemMonthlyStatService 操作系统统计
type ServerClientSystemMonthlyStatService struct {
	BaseService
}

// FindTopServerClientSystemMonthlyStats 查找前N个操作系统
func (this *ServerClientSystemMonthlyStatService) FindTopServerClientSystemMonthlyStats(ctx context.Context, req *pb.FindTopServerClientSystemMonthlyStatsRequest) (*pb.FindTopServerClientSystemMonthlyStatsResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, true)
	if err != nil {
		return nil, err
	}

	if userId > 0 {
		err = models.SharedServerDAO.CheckUserServer(nil, userId, req.ServerId)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	statList, err := stats.SharedServerClientSystemMonthlyStatDAO.ListStats(tx, req.ServerId, req.Month, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbStats := []*pb.FindTopServerClientSystemMonthlyStatsResponse_Stat{}
	for _, stat := range statList {
		pbStat := &pb.FindTopServerClientSystemMonthlyStatsResponse_Stat{
			Count:   int64(stat.Count),
			Version: stat.Version,
		}
		system, err := models.SharedClientSystemDAO.FindEnabledClientSystem(tx, int64(stat.SystemId))
		if err != nil {
			return nil, err
		}
		if system == nil {
			continue
		}
		pbStat.ClientSystem = &pb.ClientSystem{
			Id:   int64(system.Id),
			Name: system.Name,
		}
		pbStats = append(pbStats, pbStat)
	}
	return &pb.FindTopServerClientSystemMonthlyStatsResponse{Stats: pbStats}, nil
}
