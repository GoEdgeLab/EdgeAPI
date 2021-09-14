// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/nodeconfigs"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// NodeIPAddressThresholdService IP阈值相关服务
type NodeIPAddressThresholdService struct {
	BaseService
}

// CreateNodeIPAddressThreshold 创建阈值
func (this *NodeIPAddressThresholdService) CreateNodeIPAddressThreshold(ctx context.Context, req *pb.CreateNodeIPAddressThresholdRequest) (*pb.CreateNodeIPAddressThresholdResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var items = []*nodeconfigs.IPAddressThresholdItemConfig{}
	if len(req.ItemsJSON) > 0 {
		err = json.Unmarshal(req.ItemsJSON, &items)
		if err != nil {
			return nil, errors.New("decode items failed: " + err.Error())
		}
	}

	var actions = []*nodeconfigs.IPAddressThresholdActionConfig{}
	if len(req.ActionsJSON) > 0 {
		err = json.Unmarshal(req.ActionsJSON, &actions)
		if err != nil {
			return nil, errors.New("decode actions failed: " + err.Error())
		}
	}

	thresholdId, err := models.SharedNodeIPAddressThresholdDAO.CreateThreshold(tx, req.NodeIPAddressId, items, actions, 0)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNodeIPAddressThresholdResponse{NodeIPAddressThresholdId: thresholdId}, nil
}

// UpdateNodeIPAddressThreshold 修改阈值
func (this *NodeIPAddressThresholdService) UpdateNodeIPAddressThreshold(ctx context.Context, req *pb.UpdateNodeIPAddressThresholdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var items = []*nodeconfigs.IPAddressThresholdItemConfig{}
	if len(req.ItemsJSON) > 0 {
		err = json.Unmarshal(req.ItemsJSON, &items)
		if err != nil {
			return nil, errors.New("decode items failed: " + err.Error())
		}
	}

	var actions = []*nodeconfigs.IPAddressThresholdActionConfig{}
	if len(req.ActionsJSON) > 0 {
		err = json.Unmarshal(req.ActionsJSON, &actions)
		if err != nil {
			return nil, errors.New("decode actions failed: " + err.Error())
		}
	}
	err = models.SharedNodeIPAddressThresholdDAO.UpdateThreshold(tx, req.NodeIPAddressThresholdId, items, actions, -1)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// DeleteNodeIPAddressThreshold 删除阈值
func (this *NodeIPAddressThresholdService) DeleteNodeIPAddressThreshold(ctx context.Context, req *pb.DeleteNodeIPAddressThresholdRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = models.SharedNodeIPAddressThresholdDAO.DisableNodeIPAddressThreshold(tx, req.NodeIPAddressThresholdId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllEnabledNodeIPAddressThresholds 查找IP的所有阈值
func (this *NodeIPAddressThresholdService) FindAllEnabledNodeIPAddressThresholds(ctx context.Context, req *pb.FindAllEnabledNodeIPAddressThresholdsRequest) (*pb.FindAllEnabledNodeIPAddressThresholdsResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	thresholds, err := models.SharedNodeIPAddressThresholdDAO.FindAllEnabledThresholdsWithAddrId(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}

	var pbThresholds = []*pb.NodeIPAddressThreshold{}
	for _, threshold := range thresholds {
		pbThresholds = append(pbThresholds, &pb.NodeIPAddressThreshold{
			Id:          int64(threshold.Id),
			ItemsJSON:   []byte(threshold.Items),
			ActionsJSON: []byte(threshold.Actions),
		})
	}
	return &pb.FindAllEnabledNodeIPAddressThresholdsResponse{NodeIPAddressThresholds: pbThresholds}, nil
}

// CountAllEnabledNodeIPAddressThresholds 计算IP阈值的数量
func (this *NodeIPAddressThresholdService) CountAllEnabledNodeIPAddressThresholds(ctx context.Context, req *pb.CountAllEnabledNodeIPAddressThresholdsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedNodeIPAddressThresholdDAO.CountAllEnabledThresholdsWithAddrId(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// UpdateAllNodeIPAddressThresholds 批量更新阈值
func (this *NodeIPAddressThresholdService) UpdateAllNodeIPAddressThresholds(ctx context.Context, req *pb.UpdateAllNodeIPAddressThresholdsRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var thresholds = []*nodeconfigs.IPAddressThresholdConfig{}
	err = json.Unmarshal(req.NodeIPAddressThresholdsJSON, &thresholds)
	if err != nil {
		return nil, errors.New("decode thresholds failed: " + err.Error())
	}

	err = models.SharedNodeIPAddressThresholdDAO.DisableAllThresholdsWithAddrId(tx, req.NodeIPAddressId)
	if err != nil {
		return nil, err
	}
	if len(thresholds) == 0 {
		return this.Success()
	}

	var count = len(thresholds)
	for index, threshold := range thresholds {
		var order = count - index
		if threshold.Id > 0 {
			err = models.SharedNodeIPAddressThresholdDAO.UpdateThreshold(tx, threshold.Id, threshold.Items, threshold.Actions, order)
		} else {
			_, err = models.SharedNodeIPAddressThresholdDAO.CreateThreshold(tx, req.NodeIPAddressId, threshold.Items, threshold.Actions, order)
		}
		if err != nil {
			return nil, err
		}
	}

	return this.Success()
}
