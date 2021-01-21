package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	timeutil "github.com/iwind/TeaGo/utils/time"
)

// 服务统计相关服务
type ServerDailyStatService struct {
	BaseService
}

// 上传统计
func (this *ServerDailyStatService) UploadServerDailyStats(ctx context.Context, req *pb.UploadServerDailyStatsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateNode(ctx)
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
		err = models.SharedTrafficDailyStatDAO.IncreaseDailyBytes(tx, timeutil.FormatTime("Ymd", stat.CreatedAt), stat.Bytes)
		if err != nil {
			return nil, err
		}

		// 总体统计（按小时）
		err = models.SharedTrafficHourlyStatDAO.IncreaseHourlyBytes(tx, timeutil.FormatTime("YmdH", stat.CreatedAt), stat.Bytes)
		if err != nil {
			return nil, err
		}
	}

	// TODO 集群流量/节点流量

	return this.Success()
}
