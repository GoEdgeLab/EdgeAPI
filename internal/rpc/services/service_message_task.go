package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
)

// MessageTaskService 消息发送任务服务
type MessageTaskService struct {
	BaseService
}

// CreateMessageTask 创建任务
func (this *MessageTaskService) CreateMessageTask(ctx context.Context, req *pb.CreateMessageTaskRequest) (*pb.CreateMessageTaskResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	taskId, err := models.SharedMessageTaskDAO.CreateMessageTask(tx, req.RecipientId, req.InstanceId, req.User, req.Subject, req.Body, req.IsPrimary)
	if err != nil {
		return nil, err
	}
	return &pb.CreateMessageTaskResponse{MessageTaskId: taskId}, nil
}

// FindSendingMessageTasks 查找要发送的任务
func (this *MessageTaskService) FindSendingMessageTasks(ctx context.Context, req *pb.FindSendingMessageTasksRequest) (*pb.FindSendingMessageTasksResponse, error) {
	_, err := this.ValidateMonitorNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tasks, err := models.SharedMessageTaskDAO.FindSendingMessageTasks(tx, req.Size)
	if err != nil {
		return nil, err
	}
	pbTasks := []*pb.MessageTask{}
	for _, task := range tasks {
		var pbRecipient *pb.MessageRecipient
		if task.RecipientId > 0 {
			// TODO 需要缓存以提升性能
			recipient, err := models.SharedMessageRecipientDAO.FindEnabledMessageRecipient(tx, int64(task.RecipientId))
			if err != nil {
				return nil, err
			}
			if recipient == nil || recipient.IsOn == 0 {
				// 如果发送人已经删除或者禁用，则删除此消息
				err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
				if err != nil {
					return nil, err
				}
				continue
			}

			// 媒介
			// TODO 需要缓存以提升性能
			instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(recipient.InstanceId))
			if err != nil {
				return nil, err
			}
			if instance == nil || instance.IsOn == 0 {
				// 如果媒介实例已经删除或者禁用，则删除此消息
				err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
				if err != nil {
					return nil, err
				}
				continue
			}

			pbRecipient = &pb.MessageRecipient{
				Id:   int64(recipient.Id),
				User: recipient.User,
				MessageMediaInstance: &pb.MessageMediaInstance{
					Id: int64(instance.Id),
					MessageMedia: &pb.MessageMedia{
						Type: instance.MediaType,
					},
					ParamsJSON: []byte(instance.Params),
				},
			}
		} else { // 没有指定既定的接收人
			// 媒介
			// TODO 需要缓存以提升性能
			instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(task.InstanceId))
			if err != nil {
				return nil, err
			}
			if instance == nil || instance.IsOn == 0 {
				// 如果媒介实例已经删除或者禁用，则删除此消息
				err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
				if err != nil {
					return nil, err
				}
				continue
			}
			pbRecipient = &pb.MessageRecipient{
				Id: 0,
				MessageMediaInstance: &pb.MessageMediaInstance{
					Id: int64(instance.Id),
					MessageMedia: &pb.MessageMedia{
						Type: instance.MediaType,
					},
					ParamsJSON: []byte(instance.Params),
				},
			}
		}

		pbTasks = append(pbTasks, &pb.MessageTask{
			Id:               int64(task.Id),
			MessageRecipient: pbRecipient,
			User:             task.User,
			Subject:          task.Subject,
			Body:             task.Body,
			CreatedAt:        int64(task.CreatedAt),
			Status:           types.Int32(task.Status),
			SentAt:           int64(task.SentAt),
		})
	}
	return &pb.FindSendingMessageTasksResponse{MessageTasks: pbTasks}, nil
}

// UpdateMessageTaskStatus 修改任务状态
func (this *MessageTaskService) UpdateMessageTaskStatus(ctx context.Context, req *pb.UpdateMessageTaskStatusRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateMonitorNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	resultJSON := []byte{}
	if req.Result != nil {
		resultJSON, err = json.Marshal(maps.Map{
			"isOk":     req.Result.IsOk,
			"error":    req.Result.Error,
			"response": req.Result.Response,
		})
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedMessageTaskDAO.UpdateMessageTaskStatus(tx, req.MessageTaskId, int(req.Status), resultJSON)
	if err != nil {
		return nil, err
	}

	// 创建发送记录
	if (int(req.Status) == models.MessageTaskStatusSuccess || int(req.Status) == models.MessageTaskStatusFailed) && req.Result != nil {
		err = models.SharedMessageTaskLogDAO.CreateLog(tx, req.MessageTaskId, req.Result.IsOk, req.Result.Error, req.Result.Response)
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}

// DeleteMessageTask 删除消息任务
func (this *MessageTaskService) DeleteMessageTask(ctx context.Context, req *pb.DeleteMessageTaskRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageTaskDAO.DisableMessageTask(tx, req.MessageTaskId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindEnabledMessageTask 读取消息任务状态
func (this *MessageTaskService) FindEnabledMessageTask(ctx context.Context, req *pb.FindEnabledMessageTaskRequest) (*pb.FindEnabledMessageTaskResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	task, err := models.SharedMessageTaskDAO.FindEnabledMessageTask(tx, req.MessageTaskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &pb.FindEnabledMessageTaskResponse{MessageTask: nil}, nil
	}

	// TODO 需要缓存以提升性能
	var pbRecipient *pb.MessageRecipient
	if task.RecipientId > 0 {
		recipient, err := models.SharedMessageRecipientDAO.FindEnabledMessageRecipient(tx, int64(task.RecipientId))
		if err != nil {
			return nil, err
		}
		if recipient == nil || recipient.IsOn == 0 {
			// 如果发送人已经删除或者禁用，则删除此消息
			err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
			if err != nil {
				return nil, err
			}
			return &pb.FindEnabledMessageTaskResponse{MessageTask: nil}, nil
		}

		// 媒介
		// TODO 需要缓存以提升性能
		instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(recipient.InstanceId))
		if err != nil {
			return nil, err
		}
		if instance == nil || instance.IsOn == 0 {
			// 如果媒介实例已经删除或者禁用，则删除此消息
			err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
			if err != nil {
				return nil, err
			}
			return &pb.FindEnabledMessageTaskResponse{MessageTask: nil}, nil
		}

		pbRecipient = &pb.MessageRecipient{

			MessageMediaInstance: &pb.MessageMediaInstance{
				Id: int64(instance.Id),
				MessageMedia: &pb.MessageMedia{
					Type: instance.MediaType,
				},
				ParamsJSON: []byte(instance.Params),
			},
		}
	} else { // 没有指定既定的接收人
		// 媒介
		// TODO 需要缓存以提升性能
		instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(task.InstanceId))
		if err != nil {
			return nil, err
		}
		if instance == nil || instance.IsOn == 0 {
			// 如果媒介实例已经删除或者禁用，则删除此消息
			err = models.SharedMessageTaskDAO.DisableMessageTask(tx, int64(task.Id))
			if err != nil {
				return nil, err
			}
			return &pb.FindEnabledMessageTaskResponse{MessageTask: nil}, nil
		}
		pbRecipient = &pb.MessageRecipient{
			Id: 0,
			MessageMediaInstance: &pb.MessageMediaInstance{
				Id: int64(instance.Id),
				MessageMedia: &pb.MessageMedia{
					Type: instance.MediaType,
				},
				ParamsJSON: []byte(instance.Params),
			},
		}
	}

	var result = &pb.MessageTaskResult{}
	if len(task.Result) > 0 {
		err = json.Unmarshal([]byte(task.Result), result)
		if err != nil {
			return nil, err
		}
	}

	return &pb.FindEnabledMessageTaskResponse{MessageTask: &pb.MessageTask{
		Id:               int64(task.Id),
		MessageRecipient: pbRecipient,
		User:             task.User,
		Subject:          task.Subject,
		Body:             task.Body,
		CreatedAt:        int64(task.CreatedAt),
		Status:           int32(task.Status),
		SentAt:           int64(task.SentAt),
		Result:           result,
	}}, nil
}
