// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
	"github.com/iwind/TeaGo/types"
)

// FormalClientSystemService 操作系统信息库服务
type FormalClientSystemService struct {
	BaseService
}

// CreateFormalClientSystem 创建操作系统信息
func (this *FormalClientSystemService) CreateFormalClientSystem(ctx context.Context, req *pb.CreateFormalClientSystemRequest) (*pb.CreateFormalClientSystemResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	// 检查dataId是否存在
	var tx = this.NullTx()
	system, err := models.SharedFormalClientSystemDAO.FindSystemWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if system != nil {
		return nil, errors.New("dataId '" + req.DataId + "' already exists")
	}

	systemId, err := models.SharedFormalClientSystemDAO.CreateSystem(tx, req.Name, req.Codes, req.DataId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateFormalClientSystemResponse{
		FormalClientSystemId: systemId,
	}, nil
}

// CountFormalClientSystems 计算操作系统信息数量
func (this *FormalClientSystemService) CountFormalClientSystems(ctx context.Context, req *pb.CountFormalClientSystemsRequest) (*pb.RPCCountResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	count, err := models.SharedFormalClientSystemDAO.CountSystems(tx, req.Keyword)
	if err != nil {
		return nil, err
	}

	return this.SuccessCount(count)
}

// ListFormalClientSystems 列出单页操作系统信息
func (this *FormalClientSystemService) ListFormalClientSystems(ctx context.Context, req *pb.ListFormalClientSystemsRequest) (*pb.ListFormalClientSystemsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	systems, err := models.SharedFormalClientSystemDAO.ListSystems(tx, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	var pbSystems = []*pb.FormalClientSystem{}
	for _, system := range systems {
		pbSystems = append(pbSystems, &pb.FormalClientSystem{
			Id:     int64(system.Id),
			Name:   system.Name,
			Codes:  system.DecodeCodes(),
			DataId: system.DataId,
			State:  types.Int32(system.State),
		})
	}
	return &pb.ListFormalClientSystemsResponse{
		FormalClientSystems: pbSystems,
	}, nil
}

// UpdateFormalClientSystem 修改操作系统信息
func (this *FormalClientSystemService) UpdateFormalClientSystem(ctx context.Context, req *pb.UpdateFormalClientSystemRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	if len(req.DataId) == 0 {
		return nil, errors.New("invalid dataId")
	}

	var tx = this.NullTx()

	// 检查dataId是否已经被使用
	oldSystem, err := models.SharedFormalClientSystemDAO.FindSystemWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if oldSystem != nil && int64(oldSystem.Id) != req.FormalClientSystemId {
		return nil, errors.New("the dataId '" + req.DataId + "' already has been used")
	}

	err = models.SharedFormalClientSystemDAO.UpdateSystem(tx, req.FormalClientSystemId, req.Name, req.Codes, req.DataId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}

// FindFormalClientSystemWithDataId 通过dataId查询操作系统信息
func (this *FormalClientSystemService) FindFormalClientSystemWithDataId(ctx context.Context, req *pb.FindFormalClientSystemWithDataIdRequest) (*pb.FindFormalClientSystemWithDataIdResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	system, err := models.SharedFormalClientSystemDAO.FindSystemWithDataId(tx, req.DataId)
	if err != nil {
		return nil, err
	}
	if system == nil {
		return &pb.FindFormalClientSystemWithDataIdResponse{
			FormalClientSystem: nil,
		}, nil
	}

	return &pb.FindFormalClientSystemWithDataIdResponse{
		FormalClientSystem: &pb.FormalClientSystem{
			Id:     int64(system.Id),
			Name:   system.Name,
			Codes:  system.DecodeCodes(),
			DataId: system.DataId,
			State:  types.Int32(system.State),
		}}, nil
}
