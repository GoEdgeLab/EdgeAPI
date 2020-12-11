package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
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

	err = models.SharedServerDailyStatDAO.SaveStats(req.Stats)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
