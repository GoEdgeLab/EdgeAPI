package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsproviders"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNS服务商相关服务
type DNSProviderService struct {
}

// 创建服务商
func (this *DNSProviderService) CreateDNSProvider(ctx context.Context, req *pb.CreateDNSProviderRequest) (*pb.CreateDNSProviderResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	providerId, err := models.SharedDNSProviderDAO.CreateDNSProvider(req.Type, req.Name, req.ApiParamsJSON, req.RoutesJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateDNSProviderResponse{DnsProviderId: providerId}, nil
}

// 修改服务商
func (this *DNSProviderService) UpdateDNSProvider(ctx context.Context, req *pb.UpdateDNSProviderRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSProviderDAO.UpdateDNSProvider(req.DnsProviderId, req.Name, req.ApiParamsJSON, req.RoutesJSON)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 计算服务商数量
func (this *DNSProviderService) CountAllEnabledDNSProviders(ctx context.Context, req *pb.CountAllEnabledDNSProvidersRequest) (*pb.CountAllEnabledDNSProvidersResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedDNSProviderDAO.CountAllEnabledDNSProviders()
	if err != nil {
		return nil, err
	}
	return &pb.CountAllEnabledDNSProvidersResponse{Count: count}, nil
}

// 列出单页服务商信息
func (this *DNSProviderService) ListEnabledDNSProviders(ctx context.Context, req *pb.ListEnabledDNSProvidersRequest) (*pb.ListEnabledDNSProvidersResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	providers, err := models.SharedDNSProviderDAO.ListEnabledDNSProviders(req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.DNSProvider{}
	for _, provider := range providers {
		result = append(result, &pb.DNSProvider{
			Id:            int64(provider.Id),
			Name:          provider.Name,
			Type:          provider.Type,
			TypeName:      dnsproviders.FindProviderTypeName(provider.Type),
			ApiParamsJSON: []byte(provider.ApiParams),
			RoutesJSON:    []byte(provider.Routes),
			DataUpdatedAt: int64(provider.DataUpdatedAt),
		})
	}
	return &pb.ListEnabledDNSProvidersResponse{DnsProviders: result}, nil
}

// 删除服务商
func (this *DNSProviderService) DeleteDNSProvider(ctx context.Context, req *pb.DeleteDNSProviderRequest) (*pb.RPCDeleteSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSProviderDAO.DisableDNSProvider(req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCDeleteSuccess()
}

// 查找单个服务商
func (this *DNSProviderService) FindEnabledDNSProvider(ctx context.Context, req *pb.FindEnabledDNSProviderRequest) (*pb.FindEnabledDNSProviderResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	provider, err := models.SharedDNSProviderDAO.FindEnabledDNSProvider(req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &pb.FindEnabledDNSProviderResponse{DnsProvider: nil}, nil
	}

	return &pb.FindEnabledDNSProviderResponse{DnsProvider: &pb.DNSProvider{
		Id:            int64(provider.Id),
		Name:          provider.Name,
		Type:          provider.Type,
		TypeName:      dnsproviders.FindProviderTypeName(provider.Type),
		ApiParamsJSON: []byte(provider.ApiParams),
		RoutesJSON:    []byte(provider.Routes),
		DataUpdatedAt: int64(provider.DataUpdatedAt),
	}}, nil
}

// 取得所有服务商类型
func (this *DNSProviderService) FindAllDNSProviderTypes(ctx context.Context, req *pb.FindAllDNSProviderTypesRequest) (*pb.FindAllDNSProviderTypesResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSProviderType{}
	for _, t := range dnsproviders.AllProviderTypes {
		result = append(result, &pb.DNSProviderType{
			Name: t.GetString("name"),
			Code: t.GetString("code"),
		})
	}
	return &pb.FindAllDNSProviderTypesResponse{ProviderTypes: result}, nil
}
