// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/acme"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// ACMEProviderService ACME服务商
type ACMEProviderService struct {
	BaseService
}

// FindAllACMEProviders 查找所有的服务商
func (this *ACMEProviderService) FindAllACMEProviders(ctx context.Context, req *pb.FindAllACMEProvidersRequest) (*pb.FindAllACMEProvidersResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var pbProviders = []*pb.ACMEProvider{}
	for _, provider := range acme.FindAllProviders() {
		pbProviders = append(pbProviders, &pb.ACMEProvider{
			Name:           provider.Name,
			Code:           provider.Code,
			Description:    provider.Description,
			ApiURL:         provider.APIURL,
			RequireEAB:     provider.RequireEAB,
			EabDescription: provider.EABDescription,
		})
	}

	return &pb.FindAllACMEProvidersResponse{AcmeProviders: pbProviders}, nil
}

// FindACMEProviderWithCode 根据代号查找服务商
func (this *ACMEProviderService) FindACMEProviderWithCode(ctx context.Context, req *pb.FindACMEProviderWithCodeRequest) (*pb.FindACMEProviderWithCodeResponse, error) {
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	var provider = acme.FindProviderWithCode(req.AcmeProviderCode)
	if provider == nil {
		return &pb.FindACMEProviderWithCodeResponse{
			AcmeProvider: nil,
		}, nil
	}

	return &pb.FindACMEProviderWithCodeResponse{
		AcmeProvider: &pb.ACMEProvider{
			Name:           provider.Name,
			Code:           provider.Code,
			Description:    provider.Description,
			ApiURL:         provider.APIURL,
			RequireEAB:     provider.RequireEAB,
			EabDescription: provider.EABDescription,
		},
	}, nil
}
