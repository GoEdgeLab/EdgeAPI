package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// 消息接收人分组
type MessageRecipientGroupService struct {
	BaseService
}

// 创建分组
func (this *MessageRecipientGroupService) CreateMessageRecipientGroup(ctx context.Context, req *pb.CreateMessageRecipientGroupRequest) (*pb.CreateMessageRecipientGroupResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	groupId, err := models.SharedMessageRecipientGroupDAO.CreateGroup(tx, req.Name)
	if err != nil {
		return nil, err
	}

	return &pb.CreateMessageRecipientGroupResponse{MessageRecipientGroupId: groupId}, nil
}

// 修改分组
func (this *MessageRecipientGroupService) UpdateMessageRecipientGroup(ctx context.Context, req *pb.UpdateMessageRecipientGroupRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageRecipientGroupDAO.UpdateGroup(tx, req.MessageRecipientGroupId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// 查找所有可用的分组
func (this *MessageRecipientGroupService) FindAllEnabledMessageRecipientGroups(ctx context.Context, req *pb.FindAllEnabledMessageRecipientGroupsRequest) (*pb.FindAllEnabledMessageRecipientGroupsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	groups, err := models.SharedMessageRecipientGroupDAO.FindAllEnabledGroups(tx)
	if err != nil {
		return nil, err
	}
	pbGroups := []*pb.MessageRecipientGroup{}
	for _, group := range groups {
		pbGroups = append(pbGroups, &pb.MessageRecipientGroup{
			Id:   int64(group.Id),
			Name: group.Name,
			IsOn: group.IsOn,
		})
	}

	return &pb.FindAllEnabledMessageRecipientGroupsResponse{MessageRecipientGroups: pbGroups}, nil
}

// 删除分组
func (this *MessageRecipientGroupService) DeleteMessageRecipientGroup(ctx context.Context, req *pb.DeleteMessageRecipientGroupRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedMessageRecipientGroupDAO.DisableMessageRecipientGroup(tx, req.MessageRecipientGroupId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查找单个分组信息
func (this *MessageRecipientGroupService) FindEnabledMessageRecipientGroup(ctx context.Context, req *pb.FindEnabledMessageRecipientGroupRequest) (*pb.FindEnabledMessageRecipientGroupResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	group, err := models.SharedMessageRecipientGroupDAO.FindEnabledMessageRecipientGroup(tx, req.MessageRecipientGroupId)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return &pb.FindEnabledMessageRecipientGroupResponse{MessageRecipientGroup: nil}, nil
	}
	pbGroup := &pb.MessageRecipientGroup{
		Id:   int64(group.Id),
		IsOn: group.IsOn,
		Name: group.Name,
	}
	return &pb.FindEnabledMessageRecipientGroupResponse{MessageRecipientGroup: pbGroup}, nil
}
