package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models/dns"
	"github.com/TeaOSLab/EdgeAPI/internal/dnsclients"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNSProviderService DNS服务商相关服务
type DNSProviderService struct {
	BaseService
}

// CreateDNSProvider 创建服务商
func (this *DNSProviderService) CreateDNSProvider(ctx context.Context, req *pb.CreateDNSProviderRequest) (*pb.CreateDNSProviderResponse, error) {
	// 校验请求
	adminId, userId, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	providerId, err := dns.SharedDNSProviderDAO.CreateDNSProvider(tx, adminId, userId, req.Type, req.Name, req.ApiParamsJSON)
	if err != nil {
		return nil, err
	}

	return &pb.CreateDNSProviderResponse{DnsProviderId: providerId}, nil
}

// UpdateDNSProvider 修改服务商
func (this *DNSProviderService) UpdateDNSProvider(ctx context.Context, req *pb.UpdateDNSProviderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	tx := this.NullTx()

	err = dns.SharedDNSProviderDAO.UpdateDNSProvider(tx, req.DnsProviderId, req.Name, req.ApiParamsJSON)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// CountAllEnabledDNSProviders 计算服务商数量
func (this *DNSProviderService) CountAllEnabledDNSProviders(ctx context.Context, req *pb.CountAllEnabledDNSProvidersRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	count, err := dns.SharedDNSProviderDAO.CountAllEnabledDNSProviders(tx, req.AdminId, req.UserId, req.Keyword)
	if err != nil {
		return nil, err
	}
	return this.SuccessCount(count)
}

// ListEnabledDNSProviders 列出单页服务商信息
func (this *DNSProviderService) ListEnabledDNSProviders(ctx context.Context, req *pb.ListEnabledDNSProvidersRequest) (*pb.ListEnabledDNSProvidersResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	tx := this.NullTx()

	providers, err := dns.SharedDNSProviderDAO.ListEnabledDNSProviders(tx, req.AdminId, req.UserId, req.Keyword, req.Offset, req.Size)
	if err != nil {
		return nil, err
	}
	result := []*pb.DNSProvider{}
	for _, provider := range providers {
		result = append(result, &pb.DNSProvider{
			Id:            int64(provider.Id),
			Name:          provider.Name,
			Type:          provider.Type,
			TypeName:      dnsclients.FindProviderTypeName(provider.Type),
			ApiParamsJSON: provider.ApiParams,
			DataUpdatedAt: int64(provider.DataUpdatedAt),
		})
	}
	return &pb.ListEnabledDNSProvidersResponse{DnsProviders: result}, nil
}

// FindAllEnabledDNSProviders 查找所有的DNS服务商
func (this *DNSProviderService) FindAllEnabledDNSProviders(ctx context.Context, req *pb.FindAllEnabledDNSProvidersRequest) (*pb.FindAllEnabledDNSProvidersResponse, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, req.UserId)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	tx := this.NullTx()

	providers, err := dns.SharedDNSProviderDAO.FindAllEnabledDNSProviders(tx, req.AdminId, req.UserId)
	if err != nil {
		return nil, err
	}
	result := []*pb.DNSProvider{}
	for _, provider := range providers {
		result = append(result, &pb.DNSProvider{
			Id:            int64(provider.Id),
			Name:          provider.Name,
			Type:          provider.Type,
			TypeName:      dnsclients.FindProviderTypeName(provider.Type),
			ApiParamsJSON: provider.ApiParams,
			DataUpdatedAt: int64(provider.DataUpdatedAt),
		})
	}
	return &pb.FindAllEnabledDNSProvidersResponse{DnsProviders: result}, nil
}

// DeleteDNSProvider 删除服务商
func (this *DNSProviderService) DeleteDNSProvider(ctx context.Context, req *pb.DeleteDNSProviderRequest) (*pb.RPCSuccess, error) {
	// 校验请求
	_, _, err := this.ValidateAdminAndUser(ctx, 0, 0)
	if err != nil {
		return nil, err
	}

	// TODO 校验权限

	tx := this.NullTx()

	err = dns.SharedDNSProviderDAO.DisableDNSProvider(tx, req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return this.Success()
}

// FindEnabledDNSProvider 查找单个服务商
func (this *DNSProviderService) FindEnabledDNSProvider(ctx context.Context, req *pb.FindEnabledDNSProviderRequest) (*pb.FindEnabledDNSProviderResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	provider, err := dns.SharedDNSProviderDAO.FindEnabledDNSProvider(tx, req.DnsProviderId)
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
		TypeName:      dnsclients.FindProviderTypeName(provider.Type),
		ApiParamsJSON: provider.ApiParams,
		DataUpdatedAt: int64(provider.DataUpdatedAt),
	}}, nil
}

// FindAllDNSProviderTypes 取得所有服务商类型
func (this *DNSProviderService) FindAllDNSProviderTypes(ctx context.Context, req *pb.FindAllDNSProviderTypesRequest) (*pb.FindAllDNSProviderTypesResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSProviderType{}
	for _, t := range dnsclients.FindAllProviderTypes() {
		result = append(result, &pb.DNSProviderType{
			Name:        t.GetString("name"),
			Code:        t.GetString("code"),
			Description: t.GetString("description"),
		})
	}
	return &pb.FindAllDNSProviderTypesResponse{ProviderTypes: result}, nil
}

// FindAllEnabledDNSProvidersWithType 取得某个类型的所有服务商
func (this *DNSProviderService) FindAllEnabledDNSProvidersWithType(ctx context.Context, req *pb.FindAllEnabledDNSProvidersWithTypeRequest) (*pb.FindAllEnabledDNSProvidersWithTypeResponse, error) {
	// 校验请求
	_, err := this.ValidateAdmin(ctx, 0)
	if err != nil {
		return nil, err
	}

	tx := this.NullTx()

	providers, err := dns.SharedDNSProviderDAO.FindAllEnabledDNSProvidersWithType(tx, req.ProviderTypeCode)
	if err != nil {
		return nil, err
	}
	result := []*pb.DNSProvider{}
	for _, provider := range providers {
		result = append(result, &pb.DNSProvider{
			Id:       int64(provider.Id),
			Name:     provider.Name,
			Type:     provider.Type,
			TypeName: dnsclients.FindProviderTypeName(provider.Type),
		})
	}
	return &pb.FindAllEnabledDNSProvidersWithTypeResponse{DnsProviders: result}, nil
}
