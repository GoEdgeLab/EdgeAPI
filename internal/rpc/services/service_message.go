package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// MessageService 消息相关服务
type MessageService struct {
	BaseService
}

// CountUnreadMessages 计算未读消息数
func (this *MessageService) CountUnreadMessages(ctx context.Context, req *pb.CountUnreadMessagesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := models.SharedMessageDAO.CountUnreadMessages(tx, adminId, userId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListUnreadMessages 列出单页未读消息
func (this *MessageService) ListUnreadMessages(ctx context.Context, req *pb.ListUnreadMessagesRequest) (*pb.ListUnreadMessagesResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	messages, err := models.SharedMessageDAO.ListUnreadMessages(tx, adminId, userId, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Message{}
	for _, message := range messages {
		var pbCluster *pb.NodeCluster = nil
		var pbNode *pb.Node = nil

		if message.ClusterId > 0 {
			switch message.Role {
			case nodeconfigs.NodeRoleNode:
				cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(tx, int64(message.ClusterId))
				if err != nil {
					return nil, err
				}
				if cluster != nil {
					pbCluster = &pb.NodeCluster{
						Id:   int64(cluster.Id),
						Name: cluster.Name,
					}
				}
			case nodeconfigs.NodeRoleDNS:
				cluster, err := models.SharedNSClusterDAO.FindEnabledNSCluster(tx, int64(message.ClusterId))
				if err != nil {
					return nil, err
				}
				if cluster != nil {
					pbCluster = &pb.NodeCluster{
						Id:   int64(cluster.Id),
						Name: cluster.Name,
					}
				}
			}
		}

		if message.NodeId > 0 {
			switch message.Role {
			case nodeconfigs.NodeRoleNode:
				node, err := models.SharedNodeDAO.FindEnabledNode(tx, int64(message.NodeId))
				if err != nil {
					return nil, err
				}
				if node != nil {
					pbNode = &pb.Node{
						Id:   int64(node.Id),
						Name: node.Name,
					}
				}
			case nodeconfigs.NodeRoleDNS:
				node, err := models.SharedNSNodeDAO.FindEnabledNSNode(tx, int64(message.NodeId))
				if err != nil {
					return nil, err
				}
				if node != nil {
					pbNode = &pb.Node{
						Id:   int64(node.Id),
						Name: node.Name,
					}
				}
			}
		}

		result = append(result, &pb.Message{
			Id:          int64(message.Id),
			Role:        message.Role,
			Type:        message.Type,
			Body:        message.Body,
			Level:       message.Level,
			ParamsJSON:  []byte(message.Params),
			IsRead:      message.IsRead == 1,
			CreatedAt:   int64(message.CreatedAt),
			NodeCluster: pbCluster,
			Node:        pbNode,
		})
	}

	return &pb.ListUnreadMessagesResponse{Messages: result}, nil
}

// UpdateMessageRead 设置消息已读状态
func (this *MessageService) UpdateMessageRead(ctx context.Context, req *pb.UpdateMessageReadRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 校验权限
	exists, err := models.SharedMessageDAO.CheckMessageUser(tx, req.MessageId, adminId, userId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, this.PermissionError()
	}

	err = models.SharedMessageDAO.UpdateMessageRead(tx, req.MessageId, req.IsRead)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// UpdateMessagesRead 设置一组消息已读状态
func (this *MessageService) UpdateMessagesRead(ctx context.Context, req *pb.UpdateMessagesReadRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	// 校验权限
	for _, messageId := range req.MessageIds {
		exists, err := models.SharedMessageDAO.CheckMessageUser(tx, messageId, adminId, userId)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, this.PermissionError()
		}

		err = models.SharedMessageDAO.UpdateMessageRead(tx, messageId, req.IsRead)
		if err != nil {
			return nil, err
		}
	}
	return this.Success()
}

// UpdateAllMessagesRead 设置所有消息为已读
func (this *MessageService) UpdateAllMessagesRead(ctx context.Context, req *pb.UpdateAllMessagesReadRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	err = models.SharedMessageDAO.UpdateAllMessagesRead(tx, adminId, userId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
