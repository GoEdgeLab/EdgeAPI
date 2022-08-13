// Copyright 2022 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/regions"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// RegionProviderService ISP相关服务
type RegionProviderService struct {
	BaseService
}

// FindAllEnabledRegionProviders 查找所有ISP
// Deprecated
func (this *RegionProviderService) FindAllEnabledRegionProviders(ctx context.Context, req *pb.FindAllEnabledRegionProvidersRequest) (*pb.FindAllEnabledRegionProvidersResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	providers, err := regions.SharedRegionProviderDAO.FindAllEnabledProviders(tx)
	if err != nil {
		return nil, err
	}

	var pbProviders = []*pb.RegionProvider{}
	for _, provider := range providers {
		pbProviders = append(pbProviders, &pb.RegionProvider{
			Id:          int64(provider.Id),
			Name:        provider.Name,
			Codes:       provider.DecodeCodes(),
			CustomName:  provider.CustomName,
			CustomCodes: provider.DecodeCustomCodes(),
			DisplayName: provider.DisplayName(),
		})
	}

	return &pb.FindAllEnabledRegionProvidersResponse{
		RegionProviders: pbProviders,
	}, nil
}

// FindEnabledRegionProvider 查找单个ISP信息
// Deprecated
func (this *RegionProviderService) FindEnabledRegionProvider(ctx context.Context, req *pb.FindEnabledRegionProviderRequest) (*pb.FindEnabledRegionProviderResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	provider, err := regions.SharedRegionProviderDAO.FindEnabledRegionProvider(tx, req.RegionProviderId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &pb.FindEnabledRegionProviderResponse{
			RegionProvider: nil,
		}, nil
	}

	return &pb.FindEnabledRegionProviderResponse{
		RegionProvider: &pb.RegionProvider{
			Id:          int64(provider.Id),
			Name:        provider.Name,
			Codes:       provider.DecodeCodes(),
			CustomName:  provider.CustomName,
			CustomCodes: provider.DecodeCustomCodes(),
			DisplayName: provider.DisplayName(),
		},
	}, nil
}

// FindAllRegionProviders 查找所有ISP
func (this *RegionProviderService) FindAllRegionProviders(ctx context.Context, req *pb.FindAllRegionProvidersRequest) (*pb.FindAllRegionProvidersResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	providers, err := regions.SharedRegionProviderDAO.FindAllEnabledProviders(tx)
	if err != nil {
		return nil, err
	}

	var pbProviders = []*pb.RegionProvider{}
	for _, provider := range providers {
		pbProviders = append(pbProviders, &pb.RegionProvider{
			Id:          int64(provider.Id),
			Name:        provider.Name,
			Codes:       provider.DecodeCodes(),
			CustomName:  provider.CustomName,
			CustomCodes: provider.DecodeCustomCodes(),
			DisplayName: provider.DisplayName(),
		})
	}

	return &pb.FindAllRegionProvidersResponse{
		RegionProviders: pbProviders,
	}, nil
}

// FindRegionProvider 查找单个ISP信息
func (this *RegionProviderService) FindRegionProvider(ctx context.Context, req *pb.FindRegionProviderRequest) (*pb.FindRegionProviderResponse, error) {
	_, _, err := this.ValidateNodeId(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	provider, err := regions.SharedRegionProviderDAO.FindEnabledRegionProvider(tx, req.RegionProviderId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &pb.FindRegionProviderResponse{
			RegionProvider: nil,
		}, nil
	}

	return &pb.FindRegionProviderResponse{
		RegionProvider: &pb.RegionProvider{
			Id:          int64(provider.Id),
			Name:        provider.Name,
			Codes:       provider.DecodeCodes(),
			CustomName:  provider.CustomName,
			CustomCodes: provider.DecodeCustomCodes(),
			DisplayName: provider.DisplayName(),
		},
	}, nil
}

// UpdateRegionProviderCustom 修改城市定制信息
func (this *RegionProviderService) UpdateRegionProviderCustom(ctx context.Context, req *pb.UpdateRegionProviderCustomRequest) (*pb.RPCSuccess, error) {
	_, err := this.ValidateAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var tx = this.NullTx()
	err = regions.SharedRegionProviderDAO.UpdateProviderCustom(tx, req.RegionProviderId, req.CustomName, req.CustomCodes)
	if err != nil {
		return nil, err
	}
	return this.Success()
}
