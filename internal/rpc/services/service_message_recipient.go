package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// MessageRecipientService 消息接收人服务
type MessageRecipientService struct {
	BaseService
}

// CreateMessageRecipient 创建接收人
func (this *MessageRecipientService) CreateMessageRecipient(ctx context.Context, req *pb.CreateMessageRecipientRequest) (*pb.CreateMessageRecipientResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	recipientId, err := models.SharedMessageRecipientDAO.CreateRecipient(tx, req.AdminId, req.MessageMediaInstanceId, req.User, req.MessageRecipientGroupIds, req.Description, req.TimeFrom, req.TimeTo)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMessageRecipientResponse{MessageRecipientId: recipientId}, nil
}

// UpdateMessageRecipient 修改接收人
func (this *MessageRecipientService) UpdateMessageRecipient(ctx context.Context, req *pb.UpdateMessageRecipientRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageRecipientDAO.UpdateRecipient(tx, req.MessageRecipientId, req.AdminId, req.MessageMediaInstanceId, req.User, req.MessageRecipientGroupIds, req.Description, req.TimeFrom, req.TimeTo, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// DeleteMessageRecipient 删除接收人
func (this *MessageRecipientService) DeleteMessageRecipient(ctx context.Context, req *pb.DeleteMessageRecipientRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageRecipientDAO.DisableMessageRecipient(tx, req.MessageRecipientId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledMessageRecipients 计算接收人数量
func (this *MessageRecipientService) CountAllEnabledMessageRecipients(ctx context.Context, req *pb.CountAllEnabledMessageRecipientsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedMessageRecipientDAO.CountAllEnabledRecipients(tx, req.AdminId, req.MessageRecipientGroupId, req.MediaType, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListEnabledMessageRecipients 列出单页接收人
func (this *MessageRecipientService) ListEnabledMessageRecipients(ctx context.Context, req *pb.ListEnabledMessageRecipientsRequest) (*pb.ListEnabledMessageRecipientsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var cacheMap = maps.Map{}
	recipients, err := models.SharedMessageRecipientDAO.ListAllEnabledRecipients(tx, req.AdminId, req.MessageRecipientGroupId, req.MediaType, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	pbRecipients := []*pb.MessageRecipient{}
	for _, recipient := range recipients {
		// admin
		admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, int64(recipient.AdminId))
		if err != nil {
			return nil, err
		}
		if admin == nil {
			continue
		}
		pbAdmin := &pb.Admin{
			Id:       int64(admin.Id),
			Fullname: admin.Fullname,
			Username: admin.Username,
			IsOn:     admin.IsOn == 1,
		}

		// 媒介实例
		instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(recipient.InstanceId), cacheMap)
		if err != nil {
			return nil, err
		}
		if instance == nil {
			continue
		}
		pbInstance := &pb.MessageMediaInstance{
			Id:          int64(instance.Id),
			IsOn:        instance.IsOn == 1,
			Name:        instance.Name,
			Description: instance.Description,
		}

		// 分组
		pbGroups := []*pb.MessageRecipientGroup{}
		groupIds := recipient.DecodeGroupIds()
		if len(groupIds) > 0 {
			for _, groupId := range groupIds {
				group, err := models.SharedMessageRecipientGroupDAO.FindEnabledMessageRecipientGroup(tx, groupId)
				if err != nil {
					return nil, err
				}
				if group != nil {
					pbGroups = append(pbGroups, &pb.MessageRecipientGroup{
						Id:   int64(group.Id),
						Name: group.Name,
						IsOn: group.IsOn == 1,
					})
				}
			}
		}

		pbRecipients = append(pbRecipients, &pb.MessageRecipient{
			Id:                     int64(recipient.Id),
			Admin:                  pbAdmin,
			User:                   recipient.User,
			MessageMediaInstance:   pbInstance,
			IsOn:                   recipient.IsOn == 1,
			MessageRecipientGroups: pbGroups,
			Description:            recipient.Description,
			TimeFrom:               recipient.TimeFrom,
			TimeTo:                 recipient.TimeTo,
		})
	}

	return &pb.ListEnabledMessageRecipientsResponse{MessageRecipients: pbRecipients}, nil
}

// FindEnabledMessageRecipient 查找单个接收人信息
func (this *MessageRecipientService) FindEnabledMessageRecipient(ctx context.Context, req *pb.FindEnabledMessageRecipientRequest) (*pb.FindEnabledMessageRecipientResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var cacheMap = maps.Map{}
	recipient, err := models.SharedMessageRecipientDAO.FindEnabledMessageRecipient(tx, req.MessageRecipientId, cacheMap)
	if err != nil {
		return nil, err
	}
	if recipient == nil {
		return &pb.FindEnabledMessageRecipientResponse{MessageRecipient: nil}, nil
	}

	// admin
	admin, err := models.SharedAdminDAO.FindEnabledAdmin(tx, int64(recipient.AdminId))
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return &pb.FindEnabledMessageRecipientResponse{MessageRecipient: nil}, nil
	}
	pbAdmin := &pb.Admin{
		Id:       int64(admin.Id),
		Fullname: admin.Fullname,
		Username: admin.Username,
		IsOn:     admin.IsOn == 1,
	}

	// 媒介实例
	instance, err := models.SharedMessageMediaInstanceDAO.FindEnabledMessageMediaInstance(tx, int64(recipient.InstanceId), cacheMap)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return &pb.FindEnabledMessageRecipientResponse{MessageRecipient: nil}, nil
	}
	pbInstance := &pb.MessageMediaInstance{
		Id:          int64(instance.Id),
		IsOn:        instance.IsOn == 1,
		Name:        instance.Name,
		Description: instance.Description,
	}

	// 分组
	pbGroups := []*pb.MessageRecipientGroup{}
	groupIds := recipient.DecodeGroupIds()
	if len(groupIds) > 0 {
		for _, groupId := range groupIds {
			group, err := models.SharedMessageRecipientGroupDAO.FindEnabledMessageRecipientGroup(tx, groupId)
			if err != nil {
				return nil, err
			}
			if group != nil {
				pbGroups = append(pbGroups, &pb.MessageRecipientGroup{
					Id:   int64(group.Id),
					Name: group.Name,
					IsOn: group.IsOn == 1,
				})
			}
		}
	}

	return &pb.FindEnabledMessageRecipientResponse{MessageRecipient: &pb.MessageRecipient{
		Id:                     int64(recipient.Id),
		User:                   recipient.User,
		Admin:                  pbAdmin,
		MessageMediaInstance:   pbInstance,
		IsOn:                   recipient.IsOn == 1,
		MessageRecipientGroups: pbGroups,
		Description:            recipient.Description,
		TimeFrom:               recipient.TimeFrom,
		TimeTo:                 recipient.TimeTo,
	}}, nil
}
