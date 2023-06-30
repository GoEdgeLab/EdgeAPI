package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/langs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// LogService 管理员、用户或者其他系统用户日志
type LogService struct {
	BaseService
}

// CreateLog 创建日志
func (this *LogService) CreateLog(ctx context.Context, req *pb.CreateLogRequest) (*pb.CreateLogResponse, error) {
	// 校验请求
	userType, _, userId, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin, rpcutils.UserTypeUser, rpcutils.UserTypeProvider)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// i18n
	var langMessageArgs = []any{}
	if len(req.LangMessagesArgsJSON) > 0 {
		err = json.Unmarshal(req.LangMessagesArgsJSON, &langMessageArgs)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedLogDAO.CreateLog(tx, userType, userId, req.Level, req.Description, req.Action, req.Ip, langs.MessageCode(req.LangMessageCode), langMessageArgs)
	if err != nil {
		return nil, err
	}
	return &pb.CreateLogResponse{}, nil
}

// CountLogs 计算日志数量
func (this *LogService) CountLogs(ctx context.Context, req *pb.CountLogRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	count, err := models.SharedLogDAO.CountLogs(tx, req.DayFrom, req.DayTo, req.Keyword, req.UserType, req.Level)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListLogs 列出单页日志
func (this *LogService) ListLogs(ctx context.Context, req *pb.ListLogsRequest) (*pb.ListLogsResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	logs, err := models.SharedLogDAO.ListLogs(tx, req.Offset, req.Size, req.DayFrom, req.DayTo, req.Keyword, req.UserType, req.Level)
	if err != nil {
		return nil, err
	}

	result := []*pb.Log{}
	for _, log := range logs {
		userName := ""
		if log.AdminId > 0 {
			userName, err = models.SharedAdminDAO.FindAdminFullname(tx, int64(log.AdminId))
		} else if log.UserId > 0 {
			userName, err = models.SharedUserDAO.FindUserFullname(tx, int64(log.UserId))
		} else if log.ProviderId > 0 {
			userName, err = models.SharedProviderDAO.FindProviderName(tx, int64(log.ProviderId))
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

// DeleteLogPermanently 删除单条
func (this *LogService) DeleteLogPermanently(ctx context.Context, req *pb.DeleteLogPermanentlyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	var tx = this.NullTx()

	// 执行物理删除
	err = models.SharedLogDAO.DeleteLogPermanently(tx, req.LogId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteLogsPermanently 批量删除
func (this *LogService) DeleteLogsPermanently(ctx context.Context, req *pb.DeleteLogsPermanentlyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	var tx = this.NullTx()

	// 执行物理删除
	for _, logId := range req.LogIds {
		err = models.SharedLogDAO.DeleteLogPermanently(tx, logId)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// CleanLogsPermanently 清理日志
func (this *LogService) CleanLogsPermanently(ctx context.Context, req *pb.CleanLogsPermanentlyRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	var tx = this.NullTx()

	if req.ClearAll {
		err = models.SharedLogDAO.DeleteAllLogsPermanently(tx)
		if err != nil {
			return nil, err
		}
	} else if req.Days > 0 {
		err = models.SharedLogDAO.DeleteLogsPermanentlyBeforeDays(tx, int(req.Days))
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// SumLogsSize 计算日志容量大小
func (this *LogService) SumLogsSize(ctx context.Context, req *pb.SumLogsSizeRequest) (*pb.SumLogsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	size, err := models.SharedLogDAO.SumLogsSize()
	if err != nil {
		return nil, err
	}
	return &pb.SumLogsResponse{SizeBytes: size}, nil
}
