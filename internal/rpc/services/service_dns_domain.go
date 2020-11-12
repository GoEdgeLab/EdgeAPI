package services

import (
	"context"
	"github.com/TeaOSLab/EdgeAPI/internal/db/models"
	rpcutils "github.com/TeaOSLab/EdgeAPI/internal/rpc/utils"
	"github.com/TeaOSLab/EdgeCommon/pkg/rpc/pb"
)

// DNS域名相关服务
type DNSDomainService struct {
}

// 创建域名
func (this *DNSDomainService) CreateDNSDomain(ctx context.Context, req *pb.CreateDNSDomainRequest) (*pb.CreateDNSDomainResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domainId, err := models.SharedDNSDomainDAO.CreateDomain(req.DnsProviderId, req.Name)
	if err != nil {
		return nil, err
	}
	return &pb.CreateDNSDomainResponse{DnsDomainId: domainId}, nil
}

// 修改域名
func (this *DNSDomainService) UpdateDNSDomain(ctx context.Context, req *pb.UpdateDNSDomainRequest) (*pb.RPCUpdateSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSDomainDAO.UpdateDomain(req.DnsDomainId, req.Name, req.IsOn)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCUpdateSuccess()
}

// 删除域名
func (this *DNSDomainService) DeleteDNSDomain(ctx context.Context, req *pb.DeleteDNSDomainRequest) (*pb.RPCDeleteSuccess, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	err = models.SharedDNSDomainDAO.DisableDNSDomain(req.DnsDomainId)
	if err != nil {
		return nil, err
	}
	return rpcutils.RPCDeleteSuccess()
}

// 计算服务商下的域名数量
func (this *DNSDomainService) CountAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.CountAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.RPCCountResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	count, err := models.SharedDNSDomainDAO.CountAllEnabledDomainsWithProviderId(req.DnsProviderId)
	if err != nil {
		return nil, err
	}
	return &pb.RPCCountResponse{Count: count}, nil
}

// 列出服务商下的所有域名
func (this *DNSDomainService) FindAllEnabledDNSDomainsWithDNSProviderId(ctx context.Context, req *pb.FindAllEnabledDNSDomainsWithDNSProviderIdRequest) (*pb.FindAllEnabledDNSDomainsWithDNSProviderIdResponse, error) {
	// 校验请求
	_, _, err := rpcutils.ValidateRequest(ctx, rpcutils.UserTypeAdmin)
	if err != nil {
		return nil, err
	}

	domains, err := models.SharedDNSDomainDAO.FindAllEnabledDomainsWithProviderId(req.DnsProviderId)
	if err != nil {
		return nil, err
	}

	result := []*pb.DNSDomain{}
	for _, domain := range domains {
		result = append(result, &pb.DNSDomain{
			Id:            int64(domain.Id),
			Name:          domain.Name,
			IsOn:          domain.IsOn == 1,
			DataUpdatedAt: int64(domain.DataUpdatedAt),
		})
	}

	return &pb.FindAllEnabledDNSDomainsWithDNSProviderIdResponse{DnsDomains: result}, nil
}
