package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 消息相关服务
type MessageService struct {
}

// 计算未读消息数
func (this *MessageService) CountUnreadMessages(ctx context.Context, req *pb.CountUnreadMessagesRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedMessageDAO.CountUnreadMessages()
	if err != nil {
		return nil, err
	}
	return &pb.RPCCountResponse{Count: count}, nil
}

// 列出单页未读消息
func (this *MessageService) ListUnreadMessages(ctx context.Context, req *pb.ListUnreadMessagesRequest) (*pb.ListUnreadMessagesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	messages, err := models.SharedMessageDAO.ListUnreadMessages(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.Message{}
	for _, message := range messages {
		var pbCluster *pb.NodeCluster = nil
		var pbNode *pb.Node = nil

		if message.ClusterId > 0 {
			cluster, err := models.SharedNodeClusterDAO.FindEnabledNodeCluster(int64(message.ClusterId))
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

		if message.NodeId > 0 {
			node, err := models.SharedNodeDAO.FindEnabledNode(int64(message.NodeId))
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

		result = append(result, &pb.Message{
			Id:         int64(message.Id),
			Type:       message.Type,
			Body:       message.Body,
			Level:      message.Level,
			ParamsJSON: []byte(message.Params),
			IsRead:     message.IsRead == 1,
			CreatedAt:  int64(message.CreatedAt),
			Cluster:    pbCluster,
			Node:       pbNode,
		})
	}

	return &pb.ListUnreadMessagesResponse{Messages: result}, nil
}

// 设置消息已读状态
func (this *MessageService) UpdateMessageRead(ctx context.Context, req *pb.UpdateMessageReadRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedMessageDAO.UpdateMessageRead(req.MessageId, req.IsRead)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 设置一组消息已读状态
func (this *MessageService) UpdateMessagesRead(ctx context.Context, req *pb.UpdateMessagesReadRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedMessageDAO.UpdateMessagesRead(req.MessageIds, req.IsRead)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 设置所有消息为已读
func (this *MessageService) UpdateAllMessagesRead(ctx context.Context, req *pb.UpdateAllMessagesReadRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedMessageDAO.UpdateAllMessagesRead()
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}
