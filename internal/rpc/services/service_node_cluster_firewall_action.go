package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// 防火墙动作服务
type NodeClusterFirewallActionService struct {
	BaseService
}

// 创建动作
func (this *NodeClusterFirewallActionService) CreateNodeClusterFirewallAction(ctx context.Context, req *pb.CreateNodeClusterFirewallActionRequest) (*pb.NodeClusterFirewallActionResponse, error) {
	adminId, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	actionId, err := models.SharedNodeClusterFirewallActionDAO.CreateFirewallAction(tx, adminId, req.NodeClusterId, req.Name, req.EventLevel, req.Type, params)
	if err != nil {
		return nil, err
	}
	return &pb.NodeClusterFirewallActionResponse{NodeClusterFirewallActionId: actionId}, nil
}

// 修改动作
func (this *NodeClusterFirewallActionService) UpdateNodeClusterFirewallAction(ctx context.Context, req *pb.UpdateNodeClusterFirewallActionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	params := maps.Map{}
	if len(req.ParamsJSON) > 0 {
		err = json.Unmarshal(req.ParamsJSON, &params)
		if err != nil {
			return nil, err
		}
	}

	var tx = this.NullTx()
	err = models.SharedNodeClusterFirewallActionDAO.UpdateFirewallAction(tx, req.NodeClusterFirewallActionId, req.Name, req.EventLevel, req.Type, params)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 删除动作
func (this *NodeClusterFirewallActionService) DeleteNodeClusterFirewallAction(ctx context.Context, req *pb.DeleteNodeClusterFirewallActionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeClusterFirewallActionDAO.DisableFirewallAction(tx, req.NodeClusterFirewallActionId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// 查询集群的所有动作
func (this *NodeClusterFirewallActionService) FindAllEnabledNodeClusterFirewallActions(ctx context.Context, req *pb.FindAllEnabledNodeClusterFirewallActionsRequest) (*pb.FindAllEnabledNodeClusterFirewallActionsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	actions, err := models.SharedNodeClusterFirewallActionDAO.FindAllEnabledFirewallActions(tx, req.NodeClusterId)
	if err != nil {
		return nil, err
	}
	pbActions := []*pb.NodeClusterFirewallAction{}
	for _, action := range actions {
		pbActions = append(pbActions, &pb.NodeClusterFirewallAction{
			Id:            int64(action.Id),
			NodeClusterId: int64(action.ClusterId),
			Name:          action.Name,
			EventLevel:    action.EventLevel,
			Type:          action.Type,
			ParamsJSON:    []byte(action.Params),
		})
	}
	return &pb.FindAllEnabledNodeClusterFirewallActionsResponse{NodeClusterFirewallActions: pbActions}, nil
}

// 查询单个动作
func (this *NodeClusterFirewallActionService) FindEnabledNodeClusterFirewallAction(ctx context.Context, req *pb.FindEnabledNodeClusterFirewallActionRequest) (*pb.FindEnabledNodeClusterFirewallActionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	action, err := models.SharedNodeClusterFirewallActionDAO.FindEnabledFirewallAction(tx, req.NodeClusterFirewallActionId)
	if err != nil {
		return nil, err
	}
	if action == nil {
		return &pb.FindEnabledNodeClusterFirewallActionResponse{NodeClusterFirewallAction: nil}, nil
	}
	return &pb.FindEnabledNodeClusterFirewallActionResponse{NodeClusterFirewallAction: &pb.NodeClusterFirewallAction{
		Id:            int64(action.Id),
		NodeClusterId: int64(action.ClusterId),
		Name:          action.Name,
		EventLevel:    action.EventLevel,
		Type:          action.Type,
		ParamsJSON:    []byte(action.Params),
	}}, nil
}
