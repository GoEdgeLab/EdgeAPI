// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved. Official site: https://goedge.cn .

package services

import (
	"context"
	"encoding/json"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/errors"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/iplibrary"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// IPLibraryArtifactService IP库制品
type IPLibraryArtifactService struct {
	BaseService
}

// CreateIPLibraryArtifact 创建制品
func (this *IPLibraryArtifactService) CreateIPLibraryArtifact(ctx context.Context, req *pb.CreateIPLibraryArtifactRequest) (*pb.CreateIPLibraryArtifactResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	var meta = &iplibrary.Meta{}
	err = json.Unmarshal(req.MetaJSON, meta)
	if err != nil {
		return nil, errors.New("decode meta failed: " + err.Error())
	}

	artifactId, err := models.SharedIPLibraryArtifactDAO.CreateArtifact(tx, req.Name, req.FileId, 0, meta)
	if err != nil {
		return nil, err
	}
	return &pb.CreateIPLibraryArtifactResponse{IpLibraryArtifactId: artifactId}, nil
}

// UpdateIPLibraryArtifactIsPublic 使用/取消使用制品
func (this *IPLibraryArtifactService) UpdateIPLibraryArtifactIsPublic(ctx context.Context, req *pb.UpdateIPLibraryArtifactIsPublicRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedIPLibraryArtifactDAO.UpdateArtifactPublic(tx, req.IpLibraryArtifactId, req.IsPublic)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindAllIPLibraryArtifacts 查询所有制品
func (this *IPLibraryArtifactService) FindAllIPLibraryArtifacts(ctx context.Context, req *pb.FindAllIPLibraryArtifactsRequest) (*pb.FindAllIPLibraryArtifactsResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	artifacts, err := models.SharedIPLibraryArtifactDAO.FindAllArtifacts(tx)
	if err != nil {
		return nil, err
	}

	var pbArtifacts = []*pb.IPLibraryArtifact{}
	for _, artifact := range artifacts {
		pbArtifacts = append(pbArtifacts, &pb.IPLibraryArtifact{
			Id:        int64(artifact.Id),
			Name:      artifact.Name,
			FileId:    int64(artifact.FileId),
			CreatedAt: int64(artifact.CreatedAt),
			MetaJSON:  artifact.Meta,
			IsPublic:  artifact.IsPublic,
			Code:      artifact.Code,
		})
	}
	return &pb.FindAllIPLibraryArtifactsResponse{
		IpLibraryArtifacts: pbArtifacts,
	}, nil
}

// FindIPLibraryArtifact 查找当前正在使用的制品
func (this *IPLibraryArtifactService) FindIPLibraryArtifact(ctx context.Context, req *pb.FindIPLibraryArtifactRequest) (*pb.FindIPLibraryArtifactResponse, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	artifact, err := models.SharedIPLibraryArtifactDAO.FindEnabledIPLibraryArtifact(tx, req.IpLibraryArtifactId)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return &pb.FindIPLibraryArtifactResponse{
			IpLibraryArtifact: nil,
		}, nil
	}

	return &pb.FindIPLibraryArtifactResponse{
		IpLibraryArtifact: &pb.IPLibraryArtifact{
			Id:        int64(artifact.Id),
			FileId:    int64(artifact.FileId),
			CreatedAt: int64(artifact.CreatedAt),
			MetaJSON:  artifact.Meta,
			IsPublic:  artifact.IsPublic,
			Code:      artifact.Code,
		},
	}, nil
}

// FindPublicIPLibraryArtifact 查找当前正在使用的制品
func (this *IPLibraryArtifactService) FindPublicIPLibraryArtifact(ctx context.Context, req *pb.FindPublicIPLibraryArtifactRequest) (*pb.FindPublicIPLibraryArtifactResponse, error) {
	_, _, err := this.ValidateNodeId(ctx, rpcutils.UserTypeNode, rpcutils.UserTypeDNS)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	artifact, err := models.SharedIPLibraryArtifactDAO.FindPublicArtifact(tx)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return &pb.FindPublicIPLibraryArtifactResponse{
			IpLibraryArtifact: nil,
		}, nil
	}

	return &pb.FindPublicIPLibraryArtifactResponse{
		IpLibraryArtifact: &pb.IPLibraryArtifact{
			Id:        int64(artifact.Id),
			FileId:    int64(artifact.FileId),
			CreatedAt: int64(artifact.CreatedAt),
			MetaJSON:  artifact.Meta,
			IsPublic:  artifact.IsPublic,
			Code:      artifact.Code,
		},
	}, nil
}

// DeleteIPLibraryArtifact 删除制品
func (this *IPLibraryArtifactService) DeleteIPLibraryArtifact(ctx context.Context, req *pb.DeleteIPLibraryArtifactRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()

	err = models.SharedIPLibraryArtifactDAO.DisableIPLibraryArtifact(tx, req.IpLibraryArtifactId)
	if err != nil {
		return nil, err
	}

	return this.Success()
}
