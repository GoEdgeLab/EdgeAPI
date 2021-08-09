// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package nameservers

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/nameservers"
	"github.com/TeaOSLab/EdgeAPI/internal/rpc/services"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/maps"
)

// NSQuestionOptionService DNS查询选项
type NSQuestionOptionService struct {
	services.BaseService
}

// CreateNSQuestionOption 创建选项
func (this *NSQuestionOptionService) CreateNSQuestionOption(ctx context.Context, req *pb.CreateNSQuestionOptionRequest) (*pb.CreateNSQuestionOptionResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	var values = maps.Map{}
	if len(req.ValuesJSON) > 0 {
		err = json.Unmarshal(req.ValuesJSON, &values)
		if err != nil {
			return nil, err
		}
	}
	optionId, err := nameservers.SharedNSQuestionOptionDAO.CreateOption(tx, req.Name, values)
	if err != nil {
		return nil, err
	}
	return &pb.CreateNSQuestionOptionResponse{NsQuestionOptionId: optionId}, nil
}

// FindNSQuestionOption 读取选项
func (this *NSQuestionOptionService) FindNSQuestionOption(ctx context.Context, req *pb.FindNSQuestionOptionRequest) (*pb.FindNSQuestionOptionResponse, error) {
	_, err := this.ValidateNSNode(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	option, err := nameservers.SharedNSQuestionOptionDAO.FindOption(tx, req.NsQuestionOptionId)
	if err != nil {
		return nil, err
	}
	if option == nil {
		return &pb.FindNSQuestionOptionResponse{NsQuestionOption: nil}, nil
	}

	return &pb.FindNSQuestionOptionResponse{NsQuestionOption: &pb.NSQuestionOption{
		Id:         int64(option.Id),
		Name:       option.Name,
		ValuesJSON: []byte(option.Values),
	}}, nil
}

// DeleteNSQuestionOption 删除选项
func (this *NSQuestionOptionService) DeleteNSQuestionOption(ctx context.Context, req *pb.DeleteNSQuestionOptionRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = nameservers.SharedNSQuestionOptionDAO.DeleteOption(tx, req.NsQuestionOptionId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
