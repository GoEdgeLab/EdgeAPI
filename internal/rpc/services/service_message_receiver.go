package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/dbs"
	"github.com/iwind/TeaGo/maps"
)

// MessageReceiverService 消息对象接收人
type MessageReceiverService struct {
	BaseService
}

// UpdateMessageReceivers 创建接收者
func (this *MessageReceiverService) UpdateMessageReceivers(ctx context.Context, req *pb.UpdateMessageReceiversRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if len(req.Role) == 0 {
		req.Role = nodeconfigs.NodeRoleNode
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	err = this.RunTx(func(tx *dbs.Tx) error {
		err = models.SharedMessageReceiverDAO.DisableReceivers(tx, req.NodeClusterId, req.NodeId, req.ServerId)
		if err != nil {
			return err
		}

		for messageType, options := range req.RecipientOptions {
			for _, option := range options.RecipientOptions {
				_, err := models.SharedMessageReceiverDAO.CreateReceiver(tx, req.Role, req.NodeClusterId, req.NodeId, req.ServerId, messageType, params, option.MessageRecipientId, option.MessageRecipientGroupId)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindAllEnabledMessageReceivers 查找接收者
func (this *MessageReceiverService) FindAllEnabledMessageReceivers(ctx context.Context, req *pb.FindAllEnabledMessageReceiversRequest) (*pb.FindAllEnabledMessageReceiversResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if len(req.Role) == 0 {
		req.Role = nodeconfigs.NodeRoleNode
	}

	var tx = this.NullTx()
	var cacheMap = utils.NewCacheMap()
	receivers, err := models.SharedMessageReceiverDAO.FindAllEnabledReceivers(tx, req.Role, req.NodeClusterId, req.NodeId, req.ServerId, "")
	if err != nil {
		return nil, err
	}
	pbReceivers := []*pb.MessageReceiver{}
	for _, receiver := range receivers {
		var pbRecipient *pb.MessageRecipient = nil

		// 接收人
		if receiver.RecipientId > 0 {
			recipient, err := models.SharedMessageRecipientDAO.FindEnabledMessageRecipient(tx, int64(receiver.RecipientId), cacheMap)
			if err != nil {
				return nil, err
			}
			if recipient == nil {
				continue
			}

			// 管理员
			admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, int64(recipient.AdminId))
			if err != nil {
				return nil, err
			}
			if admin == nil {
				continue
			}

			// 接收人
			instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(recipient.InstanceId), cacheMap)
			if err != nil {
				return nil, err
			}
			if instance == nil {
				continue
			}

			pbRecipient = &pb.MessageRecipient{
				Id: int64(recipient.Id),
				Admin: &pb.Admin{
					Id:       int64(admin.Id),
					Fullname: admin.Fullname,
					Username: admin.Username,
					IsOn:     admin.IsOn == 1,
				},
				MessageMediaInstance: &pb.MessageMediaInstance{
					Id:   int64(instance.Id),
					Name: instance.Name,
					IsOn: instance.IsOn == 1,
				},
				IsOn:                   recipient.IsOn == 1,
				MessageRecipientGroups: nil,
				Description:            "",
				User:                   "",
			}
		}

		// 接收人分组
		var pbRecipientGroup *pb.MessageRecipientGroup = nil
		if receiver.RecipientGroupId > 0 {
			group, err := models.SharedMessageRecipientGroupDAO.FindEnabledMessageRecipientGroup(tx, int64(receiver.RecipientGroupId))
			if err != nil {
				return nil, err
			}
			if group == nil {
				continue
			}
			pbRecipientGroup = &pb.MessageRecipientGroup{
				Id:   int64(group.Id),
				Name: group.Name,
				IsOn: group.IsOn == 1,
			}
		}

		pbReceivers = append(pbReceivers, &pb.MessageReceiver{
			Id:                    int64(receiver.Id),
			ClusterId:             int64(receiver.ClusterId),
			NodeId:                int64(receiver.NodeId),
			ServerId:              int64(receiver.ServerId),
			Type:                  receiver.Type,
			ParamsJSON:            receiver.Params,
			MessageRecipient:      pbRecipient,
			MessageRecipientGroup: pbRecipientGroup,
		})
	}
	return &pb.FindAllEnabledMessageReceiversResponse{MessageReceivers: pbReceivers}, nil
}

// DeleteMessageReceiver 删除接收者
func (this *MessageReceiverService) DeleteMessageReceiver(ctx context.Context, req *pb.DeleteMessageReceiverRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageReceiverDAO.DisableMessageReceiver(tx, req.MessageReceiverId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledMessageReceivers 计算接收者数量
func (this *MessageReceiverService) CountAllEnabledMessageReceivers(ctx context.Context, req *pb.CountAllEnabledMessageReceiversRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	if len(req.Role) == 0 {
		req.Role = nodeconfigs.NodeRoleNode
	}

	var tx = this.NullTx()
	count, err := models.SharedMessageReceiverDAO.CountAllEnabledReceivers(tx, req.Role, req.NodeClusterId, req.NodeId, req.ServerId, "")
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}
