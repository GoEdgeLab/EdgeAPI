package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 管理员、用户或者其他系统用户日志
type LogService struct {
}

// 创建日志
func (this *LogService) CreateLog(ctx context.Context, req *pb.CreateLogRequest) (*pb.CreateLogResponse, error) {
	// 校验请求
	userType, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeProvider)
	if err != nil {
		return nil, err
	}

	err = models.SharedLogDAO.CreateLog(userType, userId, req.Level, req.Description, req.Action, req.Ip)
	if err != nil {
		return nil, err
	}
	return &pb.CreateLogResponse{}, nil
}

// 计算日志数量
func (this *LogService) CountLogs(ctx context.Context, req *pb.CountLogRequest) (*pb.CountLogResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedLogDAO.CountAllLogs()
	if err != nil {
		return nil, err
	}
	return &pb.CountLogResponse{Count: count}, nil
}

// 列出单页日志
func (this *LogService) ListLogs(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	logs, err := models.SharedLogDAO.ListLogs(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	result := []*pb.Log{}
	for _, log := range logs {
		userName := ""
		if log.AdminId > 0 {
			userName, err = models.SharedAdminDAO.FindAdminFullname(int64(log.AdminId))
		} else if log.UserId > 0 {
			userName, err = models.SharedUserDAO.FindUserFullname(int64(log.UserId))
		} else if log.ProviderId > 0 {
			userName, err = models.SharedProviderDAO.FindProviderName(int64(log.ProviderId))
		}

		if err != nil {
			return nil, err
		}

		result = append(result, &pb.Log{
			Id:          int64(log.Id),
			Level:       log.Level,
			Action:      log.Action,
			AdminId:     int64(log.AdminId),
			UserId:      int64(log.UserId),
			ProviderId:  int64(log.ProviderId),
			CreatedAt:   int64(log.CreatedAt),
			Type:        log.Type,
			Ip:          log.Ip,
			UserName:    userName,
			Description: log.Description,
		})
	}

	return &pb.ListLogsResponse{Logs: result}, nil
}
