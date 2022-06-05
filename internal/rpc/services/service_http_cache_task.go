// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// HTTPCacheTaskService 缓存任务管理
type HTTPCacheTaskService struct {
	BaseService
}

// CreateHTTPCacheTask 创建任务
func (this *HTTPCacheTaskService) CreateHTTPCacheTask(ctx context.Context, req *pb.CreateHTTPCacheTaskRequest) (*pb.CreateHTTPCacheTaskResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 检查操作类型
	if req.Type != models.HTTPCacheTaskTypePurge && req.Type != models.HTTPCacheTaskTypeFetch {
		return nil, errors.New("invalid type '" + req.Type + "'")
	}

	// 检查Key数量
	var clusterId int64
	if userId > 0 {
		// TODO 限制当日刷新总条数（配额）
		if len(req.Keys) > models.MaxKeysPerTask {
			return nil, errors.New("too many keys (current:" + types.String(len(req.Keys)) + ", max:" + types.String(models.MaxKeysPerTask) + ")")
		}

		clusterId, err = models.SharedUserDAO.FindUserClusterId(tx, userId)
		if err != nil {
			return nil, err
		}
	}

	// 创建任务
	taskId, err := models.SharedHTTPCacheTaskDAO.CreateTask(tx, userId, req.Type, req.KeyType, "")
	if err != nil {
		return nil, err
	}

	var countKeys = 0
	var domainMap = map[string]*models.Server{} // domain name => *Server
	for _, key := range req.Keys {
		if len(key) == 0 {
			continue
		}

		// 获取域名
		var domain = utils.ParseDomainFromKey(key)
		if len(domain) == 0 {
			continue
		}

		// 查询所在集群
		server, ok := domainMap[domain]
		if !ok {
			server, err = models.SharedServerDAO.FindEnabledServerWithDomain(tx, domain)
			if err != nil {
				return nil, err
			}
			if server == nil {
				continue
			}
			domainMap[domain] = server
		}

		// 检查用户
		if userId > 0 {
			if int64(server.UserId) != userId {
				continue
			}
		}

		var serverClusterId = int64(server.ClusterId)
		if serverClusterId == 0 {
			if clusterId > 0 {
				serverClusterId = clusterId
			} else {
				continue
			}
		}

		_, err = models.SharedHTTPCacheTaskKeyDAO.CreateKey(tx, taskId, key, req.Type, req.KeyType, serverClusterId)
		if err != nil {
			return nil, err
		}

		countKeys++
	}

	if countKeys == 0 {
		// 如果没有有效的Key，则直接完成
		err = models.SharedHTTPCacheTaskDAO.UpdateTaskStatus(tx, taskId, true, true)
	} else {
		err = models.SharedHTTPCacheTaskDAO.UpdateTaskReady(tx, taskId)
	}
	if err != nil {
		return nil, err
	}

	return &pb.CreateHTTPCacheTaskResponse{
		HttpCacheTaskId: taskId,
		CountKeys:       int64(countKeys),
	}, nil
}

// CountHTTPCacheTasks 计算任务数量
func (this *HTTPCacheTaskService) CountHTTPCacheTasks(ctx context.Context, req *pb.CountHTTPCacheTasksRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedHTTPCacheTaskDAO.CountTasks(tx, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// CountDoingHTTPCacheTasks 计算正在执行的任务数量
func (this *HTTPCacheTaskService) CountDoingHTTPCacheTasks(ctx context.Context, req *pb.CountDoingHTTPCacheTasksRequest) (*pb.RPCCountResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedHTTPCacheTaskDAO.CountDoingTasks(tx, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListHTTPCacheTasks 列出单页任务
func (this *HTTPCacheTaskService) ListHTTPCacheTasks(ctx context.Context, req *pb.ListHTTPCacheTasksRequest) (*pb.ListHTTPCacheTasksResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	tasks, err := models.SharedHTTPCacheTaskDAO.ListTasks(tx, userId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}

	var pbTasks = []*pb.HTTPCacheTask{}
	var cacheMap = utils.NewCacheMap()
	for _, task := range tasks {
		var taskId = int64(task.Id)

		// 查询所属用户
		var pbUser = &pb.User{}
		if task.UserId > 0 {
			var taskUserId = int64(task.UserId)
			if taskUserId > 0 {
				taskUser, err := models.SharedUserDAO.FindEnabledUser(tx, taskUserId, cacheMap)
				if err != nil {
					return nil, err
				}
				if taskUser == nil {
					// 找不到用户就删除
					err = models.SharedHTTPCacheTaskDAO.DisableHTTPCacheTask(tx, taskUserId)
					if err != nil {
						return nil, err
					}
				} else {
					pbUser = &pb.User{
						Id:       int64(taskUser.Id),
						Username: taskUser.Username,
						Fullname: taskUser.Fullname,
					}
				}
			}
		}

		pbTasks = append(pbTasks, &pb.HTTPCacheTask{
			Id:                taskId,
			UserId:            int64(task.UserId),
			Type:              task.Type,
			KeyType:           task.KeyType,
			CreatedAt:         int64(task.CreatedAt),
			DoneAt:            int64(task.DoneAt),
			IsDone:            task.IsDone,
			IsOk:              task.IsOk,
			Description:       task.Description,
			User:              pbUser,
			HttpCacheTaskKeys: nil,
		})
	}
	return &pb.ListHTTPCacheTasksResponse{
		HttpCacheTasks: pbTasks,
	}, nil
}

// FindEnabledHTTPCacheTask 查找单个任务
func (this *HTTPCacheTaskService) FindEnabledHTTPCacheTask(ctx context.Context, req *pb.FindEnabledHTTPCacheTaskRequest) (*pb.FindEnabledHTTPCacheTaskResponse, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPCacheTaskDAO.CheckUserTask(tx, userId, req.HttpCacheTaskId)
		if err != nil {
			return nil, err
		}
	}

	task, err := models.SharedHTTPCacheTaskDAO.FindEnabledHTTPCacheTask(tx, req.HttpCacheTaskId)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &pb.FindEnabledHTTPCacheTaskResponse{HttpCacheTask: nil}, nil
	}

	// 查询所属用户
	var pbUser = &pb.User{}
	if task.UserId > 0 {
		var taskUserId = int64(task.UserId)
		if taskUserId > 0 {
			taskUser, err := models.SharedUserDAO.FindEnabledUser(tx, taskUserId, nil)
			if err != nil {
				return nil, err
			}
			if taskUser == nil {
				// 找不到用户就删除
				err = models.SharedHTTPCacheTaskDAO.DisableHTTPCacheTask(tx, taskUserId)
				if err != nil {
					return nil, err
				}
			} else {
				pbUser = &pb.User{
					Id:       int64(taskUser.Id),
					Username: taskUser.Username,
					Fullname: taskUser.Fullname,
				}
			}
		}
	}

	// Keys
	keys, err := models.SharedHTTPCacheTaskKeyDAO.FindAllTaskKeys(tx, req.HttpCacheTaskId)
	if err != nil {
		return nil, err
	}
	var pbKeys = []*pb.HTTPCacheTaskKey{}
	for _, key := range keys {
		pbKeys = append(pbKeys, &pb.HTTPCacheTaskKey{
			Id:         int64(key.Id),
			TaskId:     int64(key.TaskId),
			Key:        key.Key,
			KeyType:    key.KeyType,
			IsDone:     key.IsDone,
			ErrorsJSON: key.Errors,
		})
	}

	return &pb.FindEnabledHTTPCacheTaskResponse{
		HttpCacheTask: &pb.HTTPCacheTask{
			Id:                int64(task.Id),
			UserId:            int64(task.UserId),
			Type:              task.Type,
			KeyType:           task.KeyType,
			CreatedAt:         int64(task.CreatedAt),
			DoneAt:            int64(task.DoneAt),
			IsDone:            task.IsDone,
			IsOk:              task.IsOk,
			User:              pbUser,
			HttpCacheTaskKeys: pbKeys,
		},
	}, nil
}

// DeleteHTTPCacheTask 删除任务
func (this *HTTPCacheTaskService) DeleteHTTPCacheTask(ctx context.Context, req *pb.DeleteHTTPCacheTaskRequest) (*pb.RPCSuccess, error) {
	_, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	if userId > 0 {
		err = models.SharedHTTPCacheTaskDAO.CheckUserTask(tx, userId, req.HttpCacheTaskId)
		if err != nil {
			return nil, err
		}
	}

	err = models.SharedHTTPCacheTaskDAO.DisableHTTPCacheTask(tx, req.HttpCacheTaskId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// ResetHTTPCacheTask 重置任务状态
// 只允许管理员重置，用于调试
func (this *HTTPCacheTaskService) ResetHTTPCacheTask(ctx context.Context, req *pb.ResetHTTPCacheTaskRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	// 重置任务
	err = models.SharedHTTPCacheTaskDAO.ResetTask(tx, req.HttpCacheTaskId)
	if err != nil {
		return nil, err
	}

	// 重置任务下的Key
	err = models.SharedHTTPCacheTaskKeyDAO.ResetCacheKeysWithTaskId(tx, req.HttpCacheTaskId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
