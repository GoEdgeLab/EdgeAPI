package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 消息发送日志相关服务
type MessageTaskLogService struct {
	BaseService
}

// 计算日志数量
func (this *MessageTaskLogService) CountMessageTaskLogs(ctx context.Context, req *pb.CountMessageTaskLogsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	count, err := models.SharedMessageTaskLogDAO.CountLogs(tx)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListMessageTaskLogs 列出当页日志
func (this *MessageTaskLogService) ListMessageTaskLogs(ctx context.Context, req *pb.ListMessageTaskLogsRequest) (*pb.ListMessageTaskLogsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}
	var tx = this.NullTx()
	logs, err := models.SharedMessageTaskLogDAO.ListLogs(tx, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	pbLogs := []*pb.MessageTaskLog{}
	for _, log := range logs {
		task, err := models.SharedMessageTaskDAO.FindEnabledMessageTask(tx, int64(log.TaskId))
		if err != nil {
			return nil, err
		}
		if task == nil {
			continue
		}

		var pbRecipient *pb.MessageRecipient
		if task.RecipientId > 0 {
			recipient, err := models.SharedMessageRecipientDAO.FindEnabledMessageRecipient(tx, int64(task.RecipientId))
			if err != nil {
				return nil, err
			}
			if recipient != nil {
				pbRecipient = &pb.MessageRecipient{
					Id:   int64(recipient.Id),
					User: recipient.User,
				}
				task.InstanceId = recipient.InstanceId
			}
		}

		instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(task.InstanceId))
		if err != nil {
			return nil, err
		}
		if instance == nil {
			continue
		}

		pbLogs = append(pbLogs, &pb.MessageTaskLog{
			Id:        int64(log.Id),
			CreatedAt: int64(log.CreatedAt),
			IsOk:      log.IsOk == 1,
			Error:     log.Error,
			Response:  log.Response,
			MessageTask: &pb.MessageTask{
				Id:               int64(task.Id),
				MessageRecipient: pbRecipient,
				MessageMediaInstance: &pb.MessageMediaInstance{
					Id:   int64(instance.Id),
					Name: instance.Name,
				},
				User:      task.User,
				Subject:   task.Subject,
				Body:      task.Body,
				CreatedAt: int64(task.CreatedAt),
				Status:    int32(task.Status),
				SentAt:    int64(task.SentAt),
				Result:    nil,
			},
		})
	}
	return &pb.ListMessageTaskLogsResponse{MessageTaskLogs: pbLogs}, nil
}
